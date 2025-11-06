package main

import (
	"log"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/adonese/cost-of-living/internal/workflow"
)

func main() {
	// Get Temporal address from env
	temporalAddress := os.Getenv("TEMPORAL_ADDRESS")
	if temporalAddress == "" {
		temporalAddress = "localhost:7233"
	}

	// Create Temporal client
	c, err := client.Dial(client.Options{
		HostPort: temporalAddress,
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()

	// Create worker
	w := worker.New(c, "cost-of-living-task-queue", worker.Options{})

	// Register workflows and activities
	w.RegisterWorkflow(workflow.HelloWorkflow)
	w.RegisterActivity(workflow.HelloActivity)

	log.Println("Worker starting...")

	// Start worker
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
