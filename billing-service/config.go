package billing

import "encore.dev/config"

type Config struct {
	// Temporal
	TemporalHost      config.String
	TemporalPort      config.Int
	TemporalNamespace config.String

	// App-level
	BillingCurrency config.String
}

var cfg = config.Load[*Config]()
