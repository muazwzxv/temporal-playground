package temporal

import (
	"context"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
)

//go:generate mockgen -source=interfaces.go -destination=mocks/mocks.go -package=mocks

// WorkflowClient defines the Temporal client operations used by handlers.
// This interface allows mocking Temporal for unit tests.
type WorkflowClient interface {
	// ExecuteWorkflow starts a new workflow execution.
	ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error)

	// SignalWorkflow sends a signal to a running workflow.
	SignalWorkflow(ctx context.Context, workflowID, runID, signalName string, arg interface{}) error

	// QueryWorkflow queries a workflow's state.
	QueryWorkflow(ctx context.Context, workflowID, runID, queryType string, args ...interface{}) (converter.EncodedValue, error)
}

// Ensure the real Temporal client satisfies our interface.
var _ WorkflowClient = (client.Client)(nil)
