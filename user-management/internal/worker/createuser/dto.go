package createuser

// Error codes for cache failure tracking
// const (
// 	ErrCodeCreateFailed = "CREATE_FAILED"
// )

type CreateUserInput struct {
	ReferenceID string
	Name        string
}

type CreateUserOutput struct {
	UserUUID string
	Name     string
	Status   string
}

type CacheSuccessInput struct {
	ReferenceID string
	UserUUID    string
	Name        string
}

type CacheFailureInput struct {
	ReferenceID  string
	ErrorCode    string
	ErrorMessage string
}
