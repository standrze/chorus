package agent

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
)

type SummarizeArgs struct {
	Text string `json:"text" description:"The text to summarize"`
}

// Summarize creates a temporary agent to summarize the given text.
// It requires an OpenAI client. Since FunctionTool functions need to match a specific signature,
// we'll return a closure that captures the client.
func NewSummarizeTool(client *openai.Client) func(SummarizeArgs) (string, error) {
	return func(args SummarizeArgs) (string, error) {
		ctx := context.Background()

		// Create a temporary agent for summarization
		summarizer := NewAgent(client,
			WithName("Summarizer"),
			WithModel("ai/gpt-oss"), // Using the standard model
			WithSystemMessage("You are a helpful assistant that summarizes text concisely."),
		)

		resp, err := summarizer.Generate(ctx, WithUserMessage(fmt.Sprintf("Please summarize the following text:\n\n%s", args.Text)))
		if err != nil {
			return "", fmt.Errorf("summarization failed: %w", err)
		}

		return resp.Choices[0].Message.Content, nil
	}
}
