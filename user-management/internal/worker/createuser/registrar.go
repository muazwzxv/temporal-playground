package createuser

import (
	"github.com/samber/do/v2"
	temporalWorker "go.temporal.io/sdk/worker"
)

type UserWorkflowRegistrar struct {
	activities *Activities
}

func NewUserWorkflowRegistrar(i do.Injector) (*UserWorkflowRegistrar, error) {
	return &UserWorkflowRegistrar{
		activities: NewUserActivities(i),
	}, nil
}

func (r *UserWorkflowRegistrar) Register(w temporalWorker.Worker) {
	w.RegisterWorkflow(CreateUserWorkflow)
	w.RegisterActivity(r.activities)
}
