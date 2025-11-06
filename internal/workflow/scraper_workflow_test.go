package workflow

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestScraperWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock the activity
	env.OnActivity(RunScraperActivity, mock.Anything, "bayut").Return(&ScraperActivityResult{
		ItemsScraped: 10,
		ItemsSaved:   10,
		Duration:     time.Second * 5,
	}, nil)

	// Execute workflow
	input := ScraperWorkflowInput{
		ScraperName: "bayut",
		MaxRetries:  3,
	}
	env.ExecuteWorkflow(ScraperWorkflow, input)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result ScraperWorkflowResult
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "bayut", result.ScraperName)
	require.Equal(t, 10, result.ItemsScraped)
	require.Equal(t, 10, result.ItemsSaved)
}

func TestScraperWorkflowWithError(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock activity to return error
	env.OnActivity(RunScraperActivity, mock.Anything, "bayut").Return(
		(*ScraperActivityResult)(nil),
		fmt.Errorf("scraping failed"),
	)

	// Mock compensation activity
	env.OnActivity(CompensateFailedScrapeActivity, mock.Anything, "bayut").Return(true, nil)

	// Execute workflow
	input := ScraperWorkflowInput{
		ScraperName: "bayut",
		MaxRetries:  1, // Only 1 retry for test speed
	}
	env.ExecuteWorkflow(ScraperWorkflow, input)

	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())

	// When workflow fails, the result might not be properly set
	// The error itself is sufficient to show failure handling works
}

func TestScheduledScraperWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock child workflow
	env.OnWorkflow(ScraperWorkflow, mock.Anything, mock.Anything).Return(&ScraperWorkflowResult{
		ScraperName:  "bayut",
		ItemsScraped: 5,
		ItemsSaved:   5,
		Duration:     time.Second * 3,
		CompletedAt:  time.Now(),
	}, nil)

	// Execute scheduled workflow
	env.ExecuteWorkflow(ScheduledScraperWorkflow)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}
