package internal

import (
	"context"
	"fmt"

	"encoding/json"
	"log"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/openai/openai-go/v3"
	chorus "github.com/standrze/chorus/pkg/agent"
	"github.com/standrze/chorus/pkg/ai"
	"github.com/standrze/chorus/pkg/tools"
)

type App struct {
	client ai.Client
	agents []*chorus.Agent
}

type AgentConfig struct {
	Name  string `mapstructure:"name"`
	Model string `mapstructure:"model"`
	//ReasoningEffort openai.ReasoningEffort `mapstructure:"reasoning_effort"`
	SystemMessage string `mapstructure:"system_message"`
}

type Config struct {
	BaseURL    string             `mapstructure:"base_url"`
	APIKey     string             `mapstructure:"api_key"`
	Agents     []AgentConfig      `mapstructure:"agents"`
	MCPServers []MCPServerConfig  `mapstructure:"mcp_servers"`
	Debug      bool               `mapstructure:"-"`
}

type MCPServerConfig struct {
	Command string   `mapstructure:"command"`
	Args    []string `mapstructure:"args"`
}

func Start(cfg *Config) error {
	app := &App{}

	app.client = ai.NewClient(cfg.BaseURL, cfg.APIKey)

	ctx := context.Background()

	// Initialize MCP Servers and fetch tools
	var mcpFuncTools []tools.FunctionTool

	for _, mcpCfg := range cfg.MCPServers {
		// Create a new client
		client := mcp.NewClient(&mcp.Implementation{
			Name:    "chorus",
			Version: "0.1.0",
		}, nil)

		// Connect to a server over stdin/stdout.
		transport := &mcp.CommandTransport{
			Command: exec.Command(mcpCfg.Command, mcpCfg.Args...),
		}

		session, err := client.Connect(ctx, transport, nil)
		if err != nil {
			log.Printf("Failed to connect to MCP server %s: %v", mcpCfg.Command, err)
			continue
		}
		// Note: we can't easily defer session.Close() here inside a loop if we want the tools to work later.
		// We probably need to keep the sessions alive.
		// For this simple implementation, we'll let them run until the app exits.
		// A proper implementation might store these sessions in the App struct and Close them on shutdown.

		listToolsRes, err := session.ListTools(ctx, nil)
		if err != nil {
			log.Printf("Failed to list tools from MCP server %s: %v", mcpCfg.Command, err)
			continue
		}

		for _, t := range listToolsRes.Tools {
			// Capture session and name for closure
			toolName := t.Name
			mcpSession := session

			// Define the wrapper function
			wrapper := func(args json.RawMessage) (string, error) {
				// unmarshal args to map[string]interface{}
				var argsMap map[string]interface{}
				if err := json.Unmarshal(args, &argsMap); err != nil {
					return "", fmt.Errorf("invalid arguments: %v", err)
				}

				callParams := &mcp.CallToolParams{
					Name:      toolName,
					Arguments: argsMap,
				}

				res, err := mcpSession.CallTool(ctx, callParams)
				if err != nil {
					return "", err
				}

				if res.IsError {
					return "", fmt.Errorf("tool execution failed")
				}

				// Combine content
				var sb string
				for _, c := range res.Content {
					if textContent, ok := c.(*mcp.TextContent); ok {
						sb += textContent.Text + "\n"
					}
				}
				return sb, nil
			}

			// Convert InputSchema (ToolInputSchema object) to openai.FunctionParameters
			// The SDK's ToolInputSchema likely has a structure we can marshal to JSON and then unmarshal?
			// Or we can just inspect it.
			// Let's assume it's compatible or we can marshal/unmarshal.

			schemaBytes, err := json.Marshal(t.InputSchema)
			if err != nil {
				log.Printf("Failed to marshal schema for tool %s: %v", t.Name, err)
				continue
			}

			var params openai.FunctionParameters
			if err := json.Unmarshal(schemaBytes, &params); err != nil {
				log.Printf("Failed to parse schema for tool %s: %v", t.Name, err)
				continue
			}

			mcpFuncTools = append(mcpFuncTools, tools.FunctionTool{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  params,
				Func:        wrapper,
			})
		}
	}

	for _, agentCfg := range cfg.Agents {
		agentOpts := []func(*chorus.Agent){
			chorus.WithReasoningEffort(openai.ReasoningEffortMedium),
			chorus.WithFunctionTools(mcpFuncTools...),
		}

		if agentCfg.Name != "" {
			agentOpts = append(agentOpts, chorus.WithName(agentCfg.Name))
		}
		if agentCfg.Model != "" {
			agentOpts = append(agentOpts, chorus.WithModel(agentCfg.Model))
		}
		if agentCfg.SystemMessage != "" {
			agentOpts = append(agentOpts, chorus.WithSystemMessage(agentCfg.SystemMessage))
		}

		agent := chorus.NewAgent(app.client, agentOpts...)
		app.agents = append(app.agents, agent)
	}

	for _, agent := range app.agents {
		fmt.Printf("Agent: %s\n", agent.Name)
		result, err := agent.Generate(ctx, chorus.WithUserMessage("Hello, how are you?"))
		if err != nil {
			return fmt.Errorf("failed to generate message for agent %s: %v", agent.Name, err)
		}
		fmt.Println(result.Choices[0].Message.Content)
	}

	return nil
}
