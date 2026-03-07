package billing

import "encore.dev/storage/sqldb"

var db = sqldb.NewDatabase("billing", sqldb.DatabaseConfig{
	Migrations: "./db/migrations",
})
