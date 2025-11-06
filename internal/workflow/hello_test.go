package workflow

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestHelloWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Register activity
	env.RegisterActivity(HelloActivity)

	// Execute workflow
	input := HelloWorkflowInput{Name: "Test User"}
	env.ExecuteWorkflow(HelloWorkflow, input)

	// Verify workflow completed successfully
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// Verify result
	var result HelloWorkflowResult
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Contains(t, result.Message, "Test User")
	require.Contains(t, result.Message, "Welcome to UAE Cost of Living")
	require.NotZero(t, result.ProcessedAt)
}

func TestHelloActivity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	// Register activity
	env.RegisterActivity(HelloActivity)

	// Execute activity
	val, err := env.ExecuteActivity(HelloActivity, "Test User")

	// Verify result
	require.NoError(t, err)

	var result string
	require.NoError(t, val.Get(&result))
	require.Equal(t, "Hello, Test User! Welcome to UAE Cost of Living", result)
}
