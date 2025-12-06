package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

// Client implements a minimal MCP client that can connect via Stdio.
type Client struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	stderr  io.ReadCloser
	scanner *bufio.Scanner
	mu      sync.Mutex
	pending map[string]chan json.RawMessage
	idSeq   int
}

type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      interface{} `json:"id,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func NewClient(command string, args []string) (*Client, error) {
	cmd := exec.Command(command, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	client := &Client{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
		scanner: bufio.NewScanner(stdout),
		pending: make(map[string]chan json.RawMessage),
	}

	go client.listen()
	go io.Copy(os.Stderr, stderr) // Pipe stderr to our stderr for debugging

	return client, nil
}

func (c *Client) listen() {
	for c.scanner.Scan() {
		line := c.scanner.Bytes()
		// Parse JSON-RPC response
		var resp JSONRPCResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			// Maybe a notification or garbage?
			continue
		}

		if resp.ID != nil {
			idStr := fmt.Sprintf("%v", resp.ID)
			c.mu.Lock()
			ch, ok := c.pending[idStr]
			if ok {
				delete(c.pending, idStr)
			}
			c.mu.Unlock()

			if ok {
				// We send the whole raw message back, or just result?
				// Let's send the full message so the caller can check errors
				ch <- line
			}
		}
	}
}

func (c *Client) Call(method string, params interface{}) (json.RawMessage, error) {
	c.mu.Lock()
	c.idSeq++
	id := c.idSeq
	idStr := fmt.Sprintf("%d", id)
	ch := make(chan json.RawMessage, 1)
	c.pending[idStr] = ch
	c.mu.Unlock()

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      id,
	}

	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// Write with newline
	if _, err := c.stdin.Write(append(b, '\n')); err != nil {
		return nil, err
	}

	// Wait for response
	respBytes := <-ch

	var resp JSONRPCResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("MCP error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	return resp.Result, nil
}

func (c *Client) Close() error {
	return c.cmd.Process.Kill()
}

// Initialize sends the initialize request
func (c *Client) Initialize() error {
	params := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"roots": map[string]interface{}{
				"listChanged": true,
			},
			"sampling": map[string]interface{}{},
		},
		"clientInfo": map[string]interface{}{
			"name":    "chorus",
			"version": "0.1.0",
		},
	}
	_, err := c.Call("initialize", params)
	return err
}

// Tool Definition from MCP spec
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"inputSchema"` // JSON Schema
}

type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}

func (c *Client) ListTools() ([]Tool, error) {
	res, err := c.Call("tools/list", nil)
	if err != nil {
		return nil, err
	}

	var result ListToolsResult
	if err := json.Unmarshal(res, &result); err != nil {
		return nil, err
	}
	return result.Tools, nil
}

type CallToolResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

func (c *Client) CallTool(name string, args map[string]interface{}) (*CallToolResult, error) {
	params := map[string]interface{}{
		"name":      name,
		"arguments": args,
	}
	res, err := c.Call("tools/call", params)
	if err != nil {
		return nil, err
	}

	var result CallToolResult
	if err := json.Unmarshal(res, &result); err != nil {
		return nil, err
	}

	if result.IsError {
		return &result, fmt.Errorf("tool execution failed")
	}

	return &result, nil
}
