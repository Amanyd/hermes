package engine

import (
	"context"
	"encoding/json"
)

type StepOutput struct {
	ActionType string          `json:"action_type"`
	OrderIndex int             `json:"order_index"`
	Output     json.RawMessage `json:"output"`
}

type ActionExecutor interface {
	Execute(ctx context.Context, config map[string]any, payload []byte, prevOutputs []StepOutput) (json.RawMessage, error)
}
