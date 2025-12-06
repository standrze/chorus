package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/standrze/chorus/pkg/tools"
)

// MockClient is hard to mock because the Agent uses a real *openai.Client.
// However, we can test the Conversation logic if we can inject behavior.
// Since we can't easily mock the OpenAI client without a complex interface,
// we will test the helper methods and the structure setup.

func TestNewConversation(t *testing.T) {
	// Setup agents
	orch := NewAgent(nil, WithName("Orchestrator"), WithRole(RoleOrchestrator))
	worker := NewAgent(nil, WithName("Worker"), WithRole(RoleAgent))

	// Success case
	conv, err := NewConversation(context.Background(), orch, worker)
	if err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	if conv.orchestrator.Name != "Orchestrator" {
		t.Errorf("Orchestrator not correctly identified")
	}
	if len(conv.agents) != 2 {
		t.Errorf("Expected 2 agents, got %d", len(conv.agents))
	}

	// Failure cases
	_, err = NewConversation(context.Background(), worker)
	if err == nil {
		t.Error("Expected error for missing orchestrator")
	}

	worker2 := NewAgent(nil, WithName("Worker"), WithRole(RoleAgent))
	_, err = NewConversation(context.Background(), orch, worker, worker2)
	if err == nil {
		t.Error("Expected error for duplicate names")
	}
}

func TestConversation_ListAgentNames(t *testing.T) {
	orch := NewAgent(nil, WithName("Orchestrator"), WithRole(RoleOrchestrator))
	worker := NewAgent(nil, WithName("Worker"), WithRole(RoleAgent))
	conv, _ := NewConversation(context.Background(), orch, worker)

	names := conv.listAgentNames()
	if !strings.Contains(names, "Orchestrator") || !strings.Contains(names, "Worker") {
		t.Errorf("listAgentNames missing agents: %s", names)
	}
}

// We can test the tools registered on the orchestrator
func TestConversation_ToolsRegistered(t *testing.T) {
	orch := NewAgent(nil, WithName("Orchestrator"), WithRole(RoleOrchestrator))
	worker := NewAgent(nil, WithName("Worker"), WithRole(RoleAgent))
	conv, _ := NewConversation(context.Background(), orch, worker)

	// Run registers the tools
	// But Run loops. We can't call it easily without mocking the API response.
	// However, we can inspect the methods that would be called.

	// Let's manually register the tool to test it (as Run does)
	orch.AddFunctionTool(tools.FunctionTool{
		Name: "DelegateTask",
		Func: conv.delegateTask,
	})

	// Test DelegateTask execution logic
	// We need to simulate the worker agent response.
	// Since Agent.Generate calls the real client, this is blocking unless we Mock the client or
	// bypass Generate.

	// For this test, we can verify that DelegateTask *tries* to find the agent.

	args := DelegateArgs{
		AgentName:    "Unknown",
		Instructions: "Do work",
	}

	_, err := conv.delegateTask(args)
	if err == nil {
		t.Error("Expected error for unknown agent")
	}

	// Note: We can't test success of DelegateTask without mocking the Worker's Generate method
	// or the OpenAI client.
}

func TestConversation_ExecuteToolCall(t *testing.T) {
	orch := NewAgent(nil, WithName("Orchestrator"), WithRole(RoleOrchestrator))
	conv, _ := NewConversation(context.Background(), orch, NewAgent(nil, WithName("W")))

	// Register a dummy tool on the orchestrator to test dispatch
	orch.AddFunctionTool(tools.FunctionTool{
		Name: "TestTool",
		Func: func(args struct{ Val string `json:"val"` }) (string, error) {
			return "Processed " + args.Val, nil
		},
	})

	// Create a mock ToolCall
	// Note: openai-go types are sometimes hard to construct manually if they have internal validation
	// but let's try.
	tc := openai.ChatCompletionMessageToolCallUnion{
		ID: "call_123",
		Function: openai.ChatCompletionMessageFunctionToolCallFunction{
			Name: "TestTool",
			Arguments: `{"val": "Input"}`,
		},
		Type: "function",
	}

	res, err := conv.executeToolCall(tc)
	if err != nil {
		t.Fatalf("executeToolCall failed: %v", err)
	}

	if res != "Processed Input" {
		t.Errorf("Unexpected result: %s", res)
	}
}
