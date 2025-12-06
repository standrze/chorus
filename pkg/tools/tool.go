package tools

import (
	"github.com/openai/openai-go/v3"
)

type FunctionTool struct {
	Name        string
	Description string
	Parameters  openai.FunctionParameters
	Type        string
	Func        any
}
