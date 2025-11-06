package main

import (
	"context"
	"fmt"
	"log"

	"go.temporal.io/sdk/client"

	"github.com/adonese/cost-of-living/internal/workflow"
)

func main() {
	// Create client
	c, err := client.Dial(client.Options{
		HostPort: "localhost:7233",
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()

	// Start workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:        "hello-workflow-example",
		TaskQueue: "cost-of-living-task-queue",
	}

	input := workflow.HelloWorkflowInput{Name: "UAE Developer"}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, workflow.HelloWorkflow, input)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	fmt.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Get workflow result
	var result workflow.HelloWorkflowResult
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable to get workflow result", err)
	}

	fmt.Printf("Workflow result: %+v\n", result)
}
