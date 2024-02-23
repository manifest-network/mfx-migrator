package store

import (
	"strings"

	"github.com/google/uuid"
)

// TODO: Move `neighborhoods/2/` to a config file
// TODO: Refactor
const (
	MigrationsEndpoint = "neighborhoods/2/migrations"
	MigrationEndpoint  = "neighborhoods/2/migrations/{uuid}"
	UpdateEndpoint     = "neighborhoods/2/migrations/{uuid}"
	AuthEndpoint       = "auth/login"
)

func GetMigrationsEndpoint() string {
	return MigrationsEndpoint
}

func GetAuthEndpoint() string {
	return AuthEndpoint
}

func GetMigrationEndpoint(uuid string) string {
	return strings.Replace(MigrationEndpoint, "{uuid}", uuid, 1)
}

func GetUpdateEndpoint(uuid uuid.UUID) string {
	return strings.Replace(UpdateEndpoint, "{uuid}", uuid.String(), 1)
}
