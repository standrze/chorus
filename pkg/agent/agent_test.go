package agent

import (
	"fmt"
	"testing"
)

func TestCallFunction(t *testing.T) {
	type EchoArgs struct {
		Message string `json:"message"`
	}

	echoTool := FunctionTool{
		Name: "Echo",
		Func: func(args EchoArgs) (string, error) {
			return "Echo: " + args.Message, nil
		},
	}

	agent := NewAgent(nil, WithFunctionTools(echoTool))

	res, err := agent.CallFunction("Echo", `{"message": "Hello"}`)
	if err != nil {
		t.Fatalf("CallFunction failed: %v", err)
	}

	if res != "Echo: Hello" {
		t.Errorf("Expected 'Echo: Hello', got '%s'", res)
	}
}

func TestCallFunction_Error(t *testing.T) {
	type FailArgs struct{}
	failTool := FunctionTool{
		Name: "Fail",
		Func: func(args FailArgs) (string, error) {
			return "", fmt.Errorf("boom")
		},
	}

	agent := NewAgent(nil, WithFunctionTools(failTool))

	_, err := agent.CallFunction("Fail", `{}`)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err.Error() != "boom" {
		t.Errorf("Expected 'boom', got '%v'", err)
	}
}

func TestCallFunction_Void(t *testing.T) {
	type VoidArgs struct{}
	voidTool := FunctionTool{
		Name: "Void",
		Func: func(args VoidArgs) error {
			return nil
		},
	}

	agent := NewAgent(nil, WithFunctionTools(voidTool))

	res, err := agent.CallFunction("Void", `{}`)
	if err != nil {
		t.Fatalf("CallFunction failed: %v", err)
	}
	if res != "Success" {
		t.Errorf("Expected 'Success', got '%s'", res)
	}
}
