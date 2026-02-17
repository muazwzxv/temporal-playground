package greeting

import (
	"context"
	"fmt"
)

// This is considered an activity in temporal context
// activity is a unit of work (uof) where that can be short or long lived ideally it's a uof that is prone to failure (network calls, database reads/writes,
// external API calls) activity can be retried by temporal depending on the configurations
func Greet(ctx context.Context, name string) (string, error) {
	return fmt.Sprintf("Hello %s", name), nil
}
