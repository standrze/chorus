package agent

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/standrze/chorus/pkg/tools"
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

	conv := &Conversation{
		ctx:          ctx,
		agents:       agentMap,
		orchestrator: orchestrator,
		maxTurns:     20,
	}

	// Inject standard tools into all agents
	// We need the client to create the Summarize tool.
	// We'll assume all agents share the same client or at least the orchestrator has one we can use.
	// In the current NewAgent setup, each agent has its own client.
	// Let's use the orchestrator's client for the Summarize tool factory.
	client := orchestrator.Client

	standardTools := []tools.FunctionTool{
		{
			Name:        "WriteToFile",
			Description: "Writes content to a file in the workspace. Overwrites if exists.",
			Func:        WriteToFile,
		},
		{
			Name:        "ReadFromFile",
			Description: "Reads content from a file in the workspace.",
			Func:        ReadFromFile,
		},
		{
			Name:        "Summarize",
			Description: "Summarizes the provided text.",
			Func:        NewSummarizeTool(client),
		},
	}

	for _, agent := range agentMap {
		for _, tool := range standardTools {
			agent.AddFunctionTool(tool)
		}
	}

	return conv, nil
}

func (c *Conversation) Interact(agentName string, instruction string) (string, error) {
	worker, exists := c.agents[agentName]
	if !exists {
		return "", fmt.Errorf("agent '%s' not found. Available agents: %s", agentName, c.listAgentNames())
	}

	worker.UserMessage(fmt.Sprintf("Task: %s", instruction))

	resp, err := worker.Generate(c.ctx)
	if err != nil {
		return "", fmt.Errorf("worker failed: %w", err)
	}

	content := resp.Choices[0].Message.Content
	worker.Messages = append(worker.Messages, openai.AssistantMessage(content))

	return content, nil
}
