package agent

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"
)

type Role string

const (
	RoleOrchestrator Role = "orchestrator"
	RoleAgent        Role = "agent"
)

type Agent struct {
	Name            string
	Role            Role
	Client          *openai.Client
	Messages        []openai.ChatCompletionMessageParamUnion
	Model           string
	Tools           []openai.ChatCompletionToolUnionParam
	ReasoningEffort openai.ReasoningEffort
	Seed            param.Opt[int64]
}

type FunctionTool struct {
	Name        string
	Description string
	Parameters  openai.FunctionParameters
	Type        string
	Func        interface{}
}

type SendOption func(*Agent)

func (a *Agent) SystemMessage(message string) {
	a.Messages = append(a.Messages, openai.SystemMessage(message))
}

func (a *Agent) UserMessage(message string) {
	a.Messages = append(a.Messages, openai.UserMessage(message))
}

func (a *Agent) Generate(ctx context.Context, options ...SendOption) (*openai.ChatCompletion, error) {
	for _, opt := range options {
		opt(a)
	}

	result, err := a.Client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:           a.Model,
		Messages:        a.Messages,
		ReasoningEffort: a.ReasoningEffort,
		Seed:            a.Seed,
		Tools:           a.Tools,
	})

	return result, err
}

func (a *Agent) AddFunctionTool(tool FunctionTool) {
	if tool.Parameters == nil && tool.Func != nil {
		schema, err := GenerateSchema(tool.Func)
		if err != nil {
			panic(fmt.Sprintf("failed to generate schema for tool %s: %v", tool.Name, err))
		}
		tool.Parameters = schema
	}

	a.Tools = append(a.Tools, openai.ChatCompletionToolUnionParam{
		OfFunction: &openai.ChatCompletionFunctionToolParam{
			Function: openai.FunctionDefinitionParam{
				Name:        tool.Name,
				Description: openai.String(tool.Description),
				Parameters:  tool.Parameters,
			},
			Type: "function",
		},
	})
}

func WithUserMessage(prompt string) SendOption {
	return func(a *Agent) {
		a.Messages = append(a.Messages, openai.UserMessage(prompt))
	}
}

func WithSystemMessage(prompt string) SendOption {
	return func(a *Agent) {
		a.Messages = append(a.Messages, openai.SystemMessage(prompt))
	}
}

func WithName(name string) func(*Agent) {
	return func(a *Agent) {
		a.Name = name
	}
}

func WithModel(model string) func(*Agent) {
	return func(a *Agent) {
		a.Model = model
	}
}

func WithReasoningEffort(reasoningEffort openai.ReasoningEffort) func(*Agent) {
	return func(a *Agent) {
		a.ReasoningEffort = reasoningEffort
	}
}

func WithSeed(seed int) func(*Agent) {
	return func(a *Agent) {
		a.Seed = openai.Int(int64(seed))
	}
}

func WithFunctionTools(tools ...FunctionTool) func(*Agent) {
	union := []openai.ChatCompletionToolUnionParam{}
	for _, tool := range tools {
		if tool.Parameters == nil && tool.Func != nil {
			schema, err := GenerateSchema(tool.Func)
			if err != nil {
				panic(fmt.Sprintf("failed to generate schema for tool %s: %v", tool.Name, err))
			}
			tool.Parameters = schema
		}

		union = append(union, openai.ChatCompletionToolUnionParam{
			OfFunction: &openai.ChatCompletionFunctionToolParam{
				Function: openai.FunctionDefinitionParam{
					Name:        tool.Name,
					Description: openai.String(tool.Description),
					Parameters:  tool.Parameters,
				},
				Type: "function",
			},
		})
	}

	return func(a *Agent) {
		a.Tools = union
	}
}

func WithRole(role Role) func(*Agent) {
	return func(a *Agent) {
		a.Role = role
	}
}

func NewAgent(client *openai.Client, options ...func(*Agent)) *Agent {
	agent := &Agent{
		Messages:        []openai.ChatCompletionMessageParamUnion{},
		Client:          client,
		Name:            GenerateAgentName(),
		Role:            RoleAgent,
		Tools:           []openai.ChatCompletionToolUnionParam{},
		Model:           "ai/gpt-oss",
		ReasoningEffort: openai.ReasoningEffortLow,
		Seed:            openai.Int(0),
	}

	for _, option := range options {
		option(agent)
	}

	return agent
}
