package worker

import "go.temporal.io/sdk/worker"

// WorkflowRegistrar is implemented by domain packages to register
// their workflows and activities with the Temporal worker.
// Each domain (user, order, etc.) should implement this interface
// to modularly register its workflows and activities.
type WorkflowRegistrar interface {
	Register(w worker.Worker)
}
