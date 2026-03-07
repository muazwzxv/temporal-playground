// Default values for local development
TemporalHost:      "127.0.0.1"
TemporalPort:      7233
TemporalNamespace: "default"
BillingCurrency:   "USD"

// Environment-specific overrides
if #Meta.Environment.Type == "production" {
    TemporalHost: "temporal.internal"
}
