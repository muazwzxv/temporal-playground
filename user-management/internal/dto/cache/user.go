package cache

type CacheSuccessInput struct {
	ReferenceID string
	UserUUID    string
	Name        string
}

type CreateUserCacheResponse struct {
	Status       string `json:"status"`
	UserUUID     string `json:"userUUID,omitempty"`
	Name         string `json:"name,omitempty"`
	ErrorCode    string `json:"errorCode,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}
