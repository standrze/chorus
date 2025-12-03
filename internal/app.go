package internal

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	chorus "github.com/standrze/chorus/pkg/agent"
)

type App struct {
	client openai.Client
	agents []*chorus.Agent
}

type AgentConfig struct {
	Name  string `mapstructure:"name"`
	Model string `mapstructure:"model"`
	//ReasoningEffort openai.ReasoningEffort `mapstructure:"reasoning_effort"`
	SystemMessage string `mapstructure:"system_message"`
}

type Config struct {
	BaseURL string        `mapstructure:"base_url"`
	APIKey  string        `mapstructure:"api_key"`
	Agents  []AgentConfig `mapstructure:"agents"`
}

func Start(cfg *Config) error {
	app := &App{}
	opts := []option.RequestOption{}

	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}

	if cfg.APIKey != "" {
		opts = append(opts, option.WithAPIKey(cfg.APIKey))
	}

	app.client = openai.NewClient(opts...)

	ctx := context.Background()

	for _, agentCfg := range cfg.Agents {
		agentOpts := []func(*chorus.Agent){
			chorus.WithReasoningEffort(openai.ReasoningEffortMedium),
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

		agent := chorus.NewAgent(&app.client, agentOpts...)
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
