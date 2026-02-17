package greeting

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.temporal.io/sdk/workflow"
)

// Worfklow in temporal
// workflow in a unit of work (uof) where it orchestrates the activities and where your application logic lives in, Temporal workflows are resilient (i dunno how yet)
// if application fails, temporal can somehow store the pre failure state and retries? i dunno how yet
func InitWorkflow(wfCtx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 10,
	}
	wfCtx = workflow.WithActivityOptions(wfCtx, ao)
	// init ctx
	ctx := context.Background()

	var result string
	err := workflow.ExecuteActivity(wfCtx, Greet, name).
		Get(wfCtx, &result)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error execurte activity, error: %+v", err))
		return "", err
	}

	return result, nil
}
