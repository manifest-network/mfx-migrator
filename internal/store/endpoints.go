package store

import (
	"strings"

	"github.com/google/uuid"
)

const (
	MigrationsEndpoint = "/migrations"
	MigrationEndpoint  = "/migrations/{uuid}"
	UpdateEndpoint     = "/migrations/{uuid}"
)

func GetMigrationsEndpoint() string {
	return MigrationsEndpoint
}

func GetMigrationEndpoint(uuid string) string {
	return strings.Replace(MigrationEndpoint, "{uuid}", uuid, 1)
}

func GetUpdateEndpoint(uuid uuid.UUID) string {
	return strings.Replace(UpdateEndpoint, "{uuid}", uuid.String(), 1)
}
