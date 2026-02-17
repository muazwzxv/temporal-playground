package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/muazwzxv/temporal-go/greeting"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// With your Activity and Workflow defined, you need a Worker to execute them. A Worker polls a Task Queue, that you configure it to poll,
// looking for work to do. Once the Worker dequeues a Workflow or Activity task from the Task Queue, it then executes that task.
func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	ctx := context.Background()

	w := worker.New(c, "local-greet-queue", worker.Options{
		MaxConcurrentActivityExecutionSize: 10,
	})

	w.RegisterWorkflow(greeting.InitWorkflow)
	w.RegisterActivity(greeting.Greet)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Unable to start worker, error: %+v", err))
		panic(err)
	}
}
