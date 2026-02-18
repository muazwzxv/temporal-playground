package user

import (
	"time"

	"github.com/samber/do/v2"
	"go.temporal.io/sdk/temporal"
	temporalWorker "go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

const (
	CreateUserWorkflowName = "CreateUserWorkflow"
)

func CreateUserWorkflow(ctx workflow.Context, input CreateUserInput) (*CreateUserOutput, error) {
	logger := workflow.GetLogger(ctx)

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        1 * time.Second,
			BackoffCoefficient:     2.0,
			MaximumInterval:        30 * time.Second,
			MaximumAttempts:        5,
			NonRetryableErrorTypes: []string{"*user.PermanentError"},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	var activities *Activities

	// Create user in database
	var output CreateUserOutput
	err := workflow.ExecuteActivity(ctx, activities.CreateUserInDB, input).Get(ctx, &output)
	if err != nil {
		logger.Error("failed to create user in database", "error", err)

		// Update cache with failure status
		cacheErr := workflow.ExecuteActivity(ctx, activities.UpdateCacheFailure, CacheFailureInput{
			ReferenceID:  input.ReferenceID,
			ErrorCode:    ErrCodeCreateFailed,
			ErrorMessage: err.Error(),
		}).Get(ctx, nil)
		if cacheErr != nil {
			logger.Warn("failed to update cache with failure status", "error", cacheErr)
		}

		return nil, err
	}

	logger.Info("user created successfully", "userUUID", output.UserUUID)

	// Update cache with success
	err = workflow.ExecuteActivity(ctx, activities.UpdateCacheSuccess, CacheSuccessInput{
		ReferenceID: input.ReferenceID,
		UserUUID:    output.UserUUID,
		Name:        output.Name,
	}).Get(ctx, nil)
	if err != nil {
		// Log but don't fail the workflow - user was created successfully
		logger.Warn("failed to update cache with success status", "error", err)
	}

	return &output, nil
}

type UserWorkflowRegistrar struct {
	activities *Activities
}

func NewUserWorkflowRegistrar(i do.Injector) (*UserWorkflowRegistrar, error) {
	return &UserWorkflowRegistrar{
		activities: NewActivities(i),
	}, nil
}

func (r *UserWorkflowRegistrar) Register(w temporalWorker.Worker) {
	w.RegisterWorkflow(CreateUserWorkflow)
	w.RegisterActivity(r.activities)
}
