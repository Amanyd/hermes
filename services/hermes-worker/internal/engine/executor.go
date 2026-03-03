package engine

import "context"

type ActionExecutor interface {
	Execute(ctx context.Context, config map[string]any, payload []byte) error
}
