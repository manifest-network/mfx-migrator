package testutils

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"

	"github.com/liftedinit/mfx-migrator/internal/store"
	"github.com/liftedinit/mfx-migrator/internal/utils"
)

const (
	DummyUUIDStr      = "5aa19d2a-4bdf-4687-a850-1804756b3f1f"
	DummyHash         = "d1e60bf3bbbe497448498f942d340b872a89046854827dc43dd703ccbf7a8c78"
	DummyManifestAddr = "manifest1jjzy5en2000728mzs3wn86a6u6jpygzajj2fg2"
	DummyCreatedDate  = "2024-03-01T16:54:02.651Z"
)

func SetupWorkItem(t *testing.T) {
	dummyUUID := uuid.MustParse(DummyUUIDStr)
	parsedCreatedDate, err := time.Parse(time.RFC3339, DummyCreatedDate)
	if err != nil {
		t.Fatal(err)
	}

	viper.Set("token-map", map[string]utils.TokenInfo{
		"dummy": {Denom: "umfx", Precision: 6},
	})

	// Some item
	item := store.WorkItem{
		Status:           2,
		CreatedDate:      &parsedCreatedDate,
		UUID:             dummyUUID,
		ManyHash:         DummyHash,
		ManifestAddress:  DummyManifestAddr,
		ManifestHash:     nil,
		ManifestDatetime: nil,
		Error:            nil,
	}

	if err := store.SaveState(&item); err != nil {
		t.Fatal(err)
	}
}
