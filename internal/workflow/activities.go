package workflow

import (
	"context"
	"fmt"
)

// HelloActivity is a simple activity that returns a greeting message
func HelloActivity(ctx context.Context, name string) (string, error) {
	return fmt.Sprintf("Hello, %s! Welcome to UAE Cost of Living", name), nil
}
