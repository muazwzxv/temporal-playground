package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/muazwzxv/temporal-go/greeting"
	"go.temporal.io/sdk/client"
)

func main() {
	ctx := context.Background()
	c, err := client.Dial(client.Options{})
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Unable to create client, err: %+v", err))
		panic(err)
	}
	defer c.Close()

	options := client.StartWorkflowOptions{
		ID:        "greeting-workflow",
		TaskQueue: "local-greet-queue",
	}

	we, err := c.ExecuteWorkflow(context.Background(), options, greeting.InitWorkflow, os.Args[1])
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Unable to execute workflow, err: %+v", err))
		panic(err)
	}
	slog.InfoContext(ctx, "Started workflow",
		"WorkflowID", we.GetID(),
		"RunID", we.GetRunID())

	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Unable get workflow result, err: %+v", err))
		panic(err)
	}
	slog.InfoContext(ctx, fmt.Sprintf("Workflow result: %v", result))
}
