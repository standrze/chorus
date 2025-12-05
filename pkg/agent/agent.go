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
	// local registry of functions for execution
	functions map[string]interface{}
}

type FunctionTool struct {
	Name        string
	Description string
	Parameters  openai.FunctionParameters
	Type        string
	Func        any
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

// CallFunction executes a registered tool function by name, unmarshaling the JSON arguments.
// It returns the result as a string or an error.
func (a *Agent) CallFunction(name string, argsJSON string) (string, error) {
	if a.functions == nil {
		return "", fmt.Errorf("no functions registered")
	}

	fn, ok := a.functions[name]
	if !ok {
		return "", fmt.Errorf("function %s not found", name)
	}

	// Reflection magic to call the function
	fnVal := reflect.ValueOf(fn)
	fnType := fnVal.Type()

	if fnType.NumIn() != 1 {
		return "", fmt.Errorf("function %s must accept exactly one argument", name)
	}

	// Create a new instance of the argument type
	argType := fnType.In(0)
	argPtr := reflect.New(argType)

	// Unmarshal JSON into the pointer
	if err := json.Unmarshal([]byte(argsJSON), argPtr.Interface()); err != nil {
		return "", fmt.Errorf("failed to unmarshal arguments for %s: %w", name, err)
	}

	// Call the function
	ret := fnVal.Call([]reflect.Value{argPtr.Elem()})

	// Handle return values
	// Expected: (string, error) or just (error) or just (string)?
	// The user earlier showed `func GenerateImage(args) (string, error)`
	// Let's support:
	// 1. (string, error)
	// 2. (string)
	// 3. (error) -> return "Success" or error message
	// 4. () -> return "Success"

	var resStr string
	var resErr error

	if len(ret) > 0 {
		// Check last return value for error
		lastVal := ret[len(ret)-1]
		if lastVal.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			if !lastVal.IsNil() {
				resErr = lastVal.Interface().(error)
			}
		}
	}

	if resErr != nil {
		return "", resErr
	}

	if len(ret) > 0 {
		firstVal := ret[0]
		if firstVal.Kind() == reflect.String {
			resStr = firstVal.String()
		}
	} else {
		resStr = "Success" // No return values
	}

	// If the first value wasn't a string (e.g. only returned error), default to success message if no error
	if resStr == "" && len(ret) == 1 && ret[0].Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		resStr = "Success"
	}

	return resStr, nil
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
	funcs := make(map[string]interface{})

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
		if a.functions == nil {
			a.functions = make(map[string]interface{})
		}
		for k, v := range funcs {
			a.functions[k] = v
		}
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
