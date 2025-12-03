package internal

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	chorus "github.com/standrze/chorus/pkg/agent"
)

type App struct {
}

type Config struct {
}

func Start() error {
	client := openai.NewClient(
		option.WithBaseURL("http://localhost:12434/engines/llama.cpp/v1"),
	)

	ctx := context.Background()

	agent := chorus.NewAgent(&client, chorus.WithName("agent"), chorus.WithModel("ai/gpt-oss"), chorus.WithReasoningEffort(openai.ReasoningEffortMedium))
	agent2 := chorus.NewAgent(&client, chorus.WithName("agent2"), chorus.WithModel("ai/gpt-oss"), chorus.WithReasoningEffort(openai.ReasoningEffortMedium))

	agent.SystemMessage(`
		You are a helpful assistant.
	`)
	agent2.SystemMessage(`
		You are a helpful assistant. 
	`)

	result, err := agent.Send(ctx, chorus.WithUserMessage("Hello, how are you?"))
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}
	fmt.Println(result.Choices[0].Message.Content)

	conversation := chorus.NewConversation(ctx, agent, agent2)
	conversation.Start("Ask a question to the user and wait for the answer.")

	return nil
}
