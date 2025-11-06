package workflow

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// HelloWorkflowInput is the input for the HelloWorkflow
type HelloWorkflowInput struct {
	Name string
}

// HelloWorkflowResult is the result of the HelloWorkflow
type HelloWorkflowResult struct {
	Message     string
	ProcessedAt time.Time
}

// HelloWorkflow is a simple hello world workflow that demonstrates Temporal basics
func HelloWorkflow(ctx workflow.Context, input HelloWorkflowInput) (*HelloWorkflowResult, error) {
	// Set activity options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 10,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Execute activity
	var result string
	err := workflow.ExecuteActivity(ctx, HelloActivity, input.Name).Get(ctx, &result)
	if err != nil {
		return nil, err
	}

	return &HelloWorkflowResult{
		Message:     result,
		ProcessedAt: workflow.Now(ctx),
	}, nil
}
