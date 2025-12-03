package agent

import (
	"context"
	"fmt"
)

type Conversation struct {
	ctx      context.Context
	agents   []*Agent
	maxTurns int
}

func NewConversation(ctx context.Context, agents ...*Agent) *Conversation {
	return &Conversation{
		ctx:      ctx,
		agents:   agents,
		maxTurns: 10,
	}
}

func (c *Conversation) Start(init string) error {
	if len(c.agents) < 2 {
		return fmt.Errorf("conversations require at least %d agents", 2)
	}

	return nil
}
