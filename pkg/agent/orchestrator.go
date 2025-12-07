package agent

import (
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/standrze/chorus/pkg/tools"
)

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

func (c *Conversation) delegateTask(args DelegateArgs) (string, error) {
	if args.AgentName == c.orchestrator.Name {
		return "", fmt.Errorf("cannot delegate to the orchestrator")
	}
	return c.Interact(args.AgentName, args.Instructions)
}

func (c *Conversation) createPlan(args PlanArgs) (string, error) {
	// Write plan to file
	planContent := "Current Plan:\n"
	for i, step := range args.Steps {
		planContent += fmt.Sprintf("%d. %s\n", i+1, step)
	}

	writeArgs := WriteArgs{
		Filename: "plan.txt", // Writing to workspace/plan.txt via WriteToFile logic implicitly or explicitly
		Content:  planContent,
	}

	// We can reuse the WriteToFile logic, but WriteToFile expects to be called as a tool.
	// We can just call the underlying logic if we want, or call WriteToFile directly since it's in the same package.
	_, err := WriteToFile(writeArgs)
	if err != nil {
		return "", fmt.Errorf("failed to save plan: %w", err)
	}

	return fmt.Sprintf("Plan created with %d steps and saved to plan.txt.", len(args.Steps)), nil
}

func (c *Conversation) listAgentNames() string {
	names := []string{}
	for n := range c.agents {
		names = append(names, n)
	}
	return strings.Join(names, ", ")
}

func (c *Conversation) Run(objective string) (string, error) {
	// Setup Orchestrator Tools
	c.orchestrator.AddFunctionTool(tools.FunctionTool{
		Name:        "DelegateTask",
		Description: "Delegate a task to a worker agent. Returns the worker's output.",
		Func:        c.delegateTask,
	})

	c.orchestrator.AddFunctionTool(tools.FunctionTool{
		Name:        "CreatePlan",
		Description: "Define the plan of execution.",
		Func:        c.createPlan,
	})

	// We can't easily break the loop with a tool call unless we handle a specific return value/error.
	// For now, let's say the Orchestrator is done when it returns a message without tool calls,
	// or we can add a explicit Finish tool.
	finished := false
	finalResult := ""

	c.orchestrator.AddFunctionTool(tools.FunctionTool{
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
		c.orchestrator.Messages = append(c.orchestrator.Messages, msg.ToParam())

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
				c.orchestrator.Messages = append(c.orchestrator.Messages, openai.ToolMessage(fmt.Sprintf("Error: %v", err), toolCall.ID))
			} else {
				c.orchestrator.Messages = append(c.orchestrator.Messages, openai.ToolMessage(res, toolCall.ID))
			}
		}
	}

	return "", fmt.Errorf("max turns reached")
}

func (c *Conversation) executeToolCall(toolCall openai.ChatCompletionMessageToolCallUnion) (string, error) {
	// Extract the function name and arguments
	name := toolCall.Function.Name
	args := toolCall.Function.Arguments

	return c.orchestrator.CallFunction(name, args)
}
