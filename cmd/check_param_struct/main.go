package main

import (
	"fmt"
	"github.com/openai/openai-go/v3"
)

func main() {
    // Check if ChatCompletionMessageToolCall exists?
    // The UnionParam must accept something.
    // Let's try to assign a struct to the UnionParam slice.

    // Logic: The API usually takes `ChatCompletionMessageToolCallParam`.
    // Maybe it's `ChatCompletionToolMessageParam`? No that's for Tool outputs.
    // `ChatCompletionAssistantMessageParam` has `ToolCalls`.

    // Let's print the name of the type that satisfies the interface?
    // Or just look for likely candidates.

    // We know `ChatCompletionMessageToolCallUnion` exists (response).
    // Maybe `ChatCompletionMessageToolCallParam` is the name but I got the package wrong?
    // No, it's `openai`.

    // Let's try `ChatCompletionMessageToolCall`.
    var x openai.ChatCompletionMessageToolCall
    fmt.Printf("%T\n", x)
}
