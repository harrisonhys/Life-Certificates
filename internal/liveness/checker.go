package liveness

import (
	"context"
)

// Checker defines the behaviour for liveness detection providers.
type Checker interface {
	Evaluate(ctx context.Context, image []byte) (passed bool, reason string, err error)
}

// NoopChecker is a simple implementation that always returns success.
type NoopChecker struct {
	Enabled bool
}

// Evaluate returns true when enabled or signals REVIEW when disabled.
func (n NoopChecker) Evaluate(_ context.Context, _ []byte) (bool, string, error) {
	if !n.Enabled {
		return false, "liveness_disabled", nil
	}
	return true, "ok", nil
}
