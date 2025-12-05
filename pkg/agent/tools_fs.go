package agent

import (
	"fmt"
	"os"
	"path/filepath"
)

// Ensure workspace exists
const workspaceDir = "workspace"

func ensureWorkspace() error {
	return os.MkdirAll(workspaceDir, 0755)
}

type WriteArgs struct {
	Filename string `json:"filename" description:"The name of the file to write to"`
	Content  string `json:"content" description:"The content to write to the file"`
}

type ReadArgs struct {
	Filename string `json:"filename" description:"The name of the file to read"`
}

func WriteToFile(args WriteArgs) (string, error) {
	if err := ensureWorkspace(); err != nil {
		return "", fmt.Errorf("failed to create workspace: %w", err)
	}

	path := filepath.Join(workspaceDir, args.Filename)
	if err := os.WriteFile(path, []byte(args.Content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully wrote to %s", args.Filename), nil
}

func ReadFromFile(args ReadArgs) (string, error) {
	if err := ensureWorkspace(); err != nil {
		return "", fmt.Errorf("failed to create workspace: %w", err)
	}

	path := filepath.Join(workspaceDir, args.Filename)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file not found: %s", args.Filename)
		}
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(data), nil
}
