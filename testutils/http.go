package testutils

import (
	"fmt"
)

const (
	Uuidv4Regex  = "[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89aAbB][a-f0-9]{3}-[a-f0-9]{12}"
	RootUrl      = "http://fakeurl:3001/api/v1/"
	Neighborhood = "1"
)

var (
	DefaultMigrationsUrl = RootUrl + fmt.Sprintf("neighborhoods/%s/migrations/", "0")
	DefaultMigrationUrl  = DefaultMigrationsUrl + Uuidv4Regex

	DefaultTransactionUrl = RootUrl + fmt.Sprintf("neighborhoods/%s/transactions/", "0")
	DefaultClaimUrl       = RootUrl + fmt.Sprintf("neighborhoods/%s/migrations/claim/", "0")

	ClaimUrl     = RootUrl + fmt.Sprintf("neighborhoods/%s/migrations/claim/", Neighborhood)
	ClaimUuidUrl = ClaimUrl + Uuidv4Regex
	LoginUrl     = RootUrl + "auth/login"
)
