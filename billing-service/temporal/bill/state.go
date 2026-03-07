package bill

// billWorkflowState holds the mutable state of the bill workflow.
// Encapsulating state in a struct makes it easier to:
// - Return state in query handlers
// - Reason about state transitions
// - Test state management in isolation
type billWorkflowState struct {
	Status     string
	TotalCents int64
	ItemCount  int
}
