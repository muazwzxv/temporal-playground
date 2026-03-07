package temporal

import (
	"fmt"

	"go.temporal.io/sdk/client"
)

func NewClient(host string, port int, namespace string) (client.Client, error) {
	c, err := client.Dial(client.Options{
		HostPort:  fmt.Sprintf("%s:%d", host, port),
		Namespace: namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("temporal client dial: %w", err)
	}

	return c, nil
}
