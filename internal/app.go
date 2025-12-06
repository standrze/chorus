package internal

import (
	"context"
	"fmt"

	"encoding/json"
	"log"

	"github.com/openai/openai-go/v3"
	chorus "github.com/standrze/chorus/pkg/agent"
	"github.com/standrze/chorus/pkg/ai"
	"github.com/standrze/chorus/pkg/mcp"
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
		client, err := mcp.NewClient(mcpCfg.Command, mcpCfg.Args)
		if err != nil {
			log.Printf("Failed to connect to MCP server %s: %v", mcpCfg.Command, err)
			continue
		}

		if err := client.Initialize(); err != nil {
			log.Printf("Failed to initialize MCP server %s: %v", mcpCfg.Command, err)
			continue
		}

		mcpToolsList, err := client.ListTools()
		if err != nil {
			log.Printf("Failed to list tools from MCP server %s: %v", mcpCfg.Command, err)
			continue
		}

		for _, t := range mcpToolsList {
			// Capture client and name for closure
			toolName := t.Name
			mcpClient := client

			// Define the wrapper function
			wrapper := func(args json.RawMessage) (string, error) {
				// unmarshal args to map[string]interface{}
				var argsMap map[string]interface{}
				if err := json.Unmarshal(args, &argsMap); err != nil {
					return "", fmt.Errorf("invalid arguments: %v", err)
				}
				res, err := mcpClient.CallTool(toolName, argsMap)
				if err != nil {
					return "", err
				}
				// Combine content
				var sb string
				for _, c := range res.Content {
					if c.Type == "text" {
						sb += c.Text + "\n"
					}
				}
				return sb, nil
			}

			// Convert InputSchema (json.RawMessage) to openai.FunctionParameters (map[string]interface{})
			var params openai.FunctionParameters
			if err := json.Unmarshal(t.InputSchema, &params); err != nil {
				log.Printf("Failed to parse schema for tool %s: %v", t.Name, err)
				continue
			}

			mcpFuncTools = append(mcpFuncTools, tools.FunctionTool{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  params,
				Func:        wrapper, // The agent.CallFunction logic needs to handle this.
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
