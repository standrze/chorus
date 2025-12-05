package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"
)

type Conversation struct {
	ctx          context.Context
	agents       map[string]*Agent
	orchestrator *Agent
	maxTurns     int
}

func NewConversation(ctx context.Context, agents ...*Agent) (*Conversation, error) {
	agentMap := make(map[string]*Agent)
	var orchestrator *Agent

	for _, a := range agents {
		if _, exists := agentMap[a.Name]; exists {
			return nil, fmt.Errorf("duplicate agent name: %s", a.Name)
		}
		agentMap[a.Name] = a
		if a.Role == RoleOrchestrator {
			if orchestrator != nil {
				return nil, fmt.Errorf("multiple orchestrators found")
			}
			orchestrator = a
		}
	}

	if orchestrator == nil {
		return nil, fmt.Errorf("no orchestrator found")
	}

	if len(agentMap) < 2 {
		return nil, fmt.Errorf("conversation requires at least 2 agents (1 orchestrator + 1 worker)")
	}

	return &Conversation{
		ctx:          ctx,
		agents:       agentMap,
		orchestrator: orchestrator,
		maxTurns:     20,
	}, nil
}

type DelegateArgs struct {
	AgentName    string `json:"agent_name" description:"The name of the agent to delegate to"`
	Instructions string `json:"instructions" description:"The task instructions for the agent"`
}

type PlanArgs struct {
	Steps []string `json:"steps" description:"The list of steps for the plan"`
}

type FinishArgs struct {
	Result string `json:"result" description:"The final result of the conversation"`
}

func (c *Conversation) Run(objective string) (string, error) {
	// Setup Orchestrator Tools
	c.orchestrator.AddFunctionTool(FunctionTool{
		Name:        "DelegateTask",
		Description: "Delegate a task to a worker agent. Returns the worker's output.",
		Func:        c.delegateTask,
	})

	c.orchestrator.AddFunctionTool(FunctionTool{
		Name:        "CreatePlan",
		Description: "Define the plan of execution.",
		Func:        c.createPlan,
	})

	// We can't easily break the loop with a tool call unless we handle a specific return value/error.
	// For now, let's say the Orchestrator is done when it returns a message without tool calls,
	// or we can add a explicit Finish tool.
	finished := false
	finalResult := ""

	c.orchestrator.AddFunctionTool(FunctionTool{
		Name:        "Finish",
		Description: "Call this when the objective is met.",
		Func: func(args FinishArgs) (string, error) {
			finished = true
			finalResult = args.Result
			return "Conversation finished.", nil
		},
	})

	// Initial Prompt
	c.orchestrator.UserMessage(fmt.Sprintf("Objective: %s", objective))

	for i := 0; i < c.maxTurns; i++ {
		if finished {
			return finalResult, nil
		}

		resp, err := c.orchestrator.Generate(c.ctx)
		if err != nil {
			return "", fmt.Errorf("orchestrator generation failed: %w", err)
		}

		choice := resp.Choices[0]
		msg := choice.Message

		// Add assistant message to history
		// We must explicitly include ToolCalls for the API context to be valid for subsequent ToolMessages
		assistantMsg := openai.AssistantMessage(msg.Content)
		if len(msg.ToolCalls) > 0 {
			toolCalls := []openai.ChatCompletionMessageToolCallUnionParam{}
			for _, tc := range msg.ToolCalls {
				toolCalls = append(toolCalls, tc.ToParam())
			}
			assistantMsg.OfAssistant.ToolCalls = toolCalls
		}
		c.orchestrator.Messages = append(c.orchestrator.Messages, assistantMsg)

		if len(msg.ToolCalls) == 0 {
			// If Orchestrator just talks, maybe we should continue or stop?
			// The user wants strict guidance. Let's assume if it stops calling tools, it might be waiting for user input,
			// but here we are automated. Let's assume it's just a comment.
			// Or maybe we treat a non-tool message as the final result if 'Finish' wasn't called?
			// Let's force it to use 'Finish'.
			continue
		}

		// Handle Tool Calls
		for _, toolCall := range msg.ToolCalls {
			res, err := c.executeToolCall(toolCall)
			if err != nil {
				// Feed error back to agent
				c.orchestrator.Messages = append(c.orchestrator.Messages, openai.ToolMessage(toolCall.ID, fmt.Sprintf("Error: %v", err)))
			} else {
				c.orchestrator.Messages = append(c.orchestrator.Messages, openai.ToolMessage(toolCall.ID, res))
			}
		}
	}

	return "", fmt.Errorf("max turns reached")
}

func (c *Conversation) delegateTask(args DelegateArgs) (string, error) {
	worker, exists := c.agents[args.AgentName]
	if !exists {
		return "", fmt.Errorf("agent '%s' not found. Available agents: %s", args.AgentName, c.listAgentNames())
	}

	if worker.Role == RoleOrchestrator {
		return "", fmt.Errorf("cannot delegate to the orchestrator")
	}

	// Scope: Worker gets fresh context + instructions
	// We might want to preserve the worker's memory across multiple calls *within* the conversation?
	// "only what is relevant to them".
	// If I ask the Writer to write Ch1, then later Ch2, it should probably remember Ch1.
	// So I will append to the worker's existing messages.

	worker.UserMessage(fmt.Sprintf("Task: %s", args.Instructions))

	resp, err := worker.Generate(c.ctx)
	if err != nil {
		return "", fmt.Errorf("worker failed: %w", err)
	}

	content := resp.Choices[0].Message.Content
	worker.Messages = append(worker.Messages, openai.AssistantMessage(content))

	return content, nil
}

func (c *Conversation) createPlan(args PlanArgs) (string, error) {
	// Just verify the plan format or log it.
	return fmt.Sprintf("Plan created with %d steps.", len(args.Steps)), nil
}

func (c *Conversation) listAgentNames() string {
	names := []string{}
	for n := range c.agents {
		names = append(names, n)
	}
	return strings.Join(names, ", ")
}

func (c *Conversation) executeToolCall(toolCall openai.ChatCompletionMessageToolCallUnion) (string, error) {
	// Extract the function name and arguments
	name := toolCall.Function.Name
	args := toolCall.Function.Arguments

	return c.orchestrator.CallFunction(name, args)
}
