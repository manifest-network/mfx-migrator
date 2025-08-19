package utils_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/manifest-network/mfx-migrator/internal/utils"
)

func TestPtr(t *testing.T) {
	t.Parallel()

	t.Run("EqualStringPtr", func(t *testing.T) {
		t.Parallel()

		foo := "foo"
		bar := "bar"
		empty := ""

		tt := []struct {
			name     string
			a        *string
			b        *string
			expected bool
		}{
			{name: "both nil", a: nil, b: nil, expected: true},
			{name: "a nil", a: nil, b: &foo, expected: false},
			{name: "b nil", a: &foo, b: nil, expected: false},
			{name: "both empty", a: &empty, b: &empty, expected: true},
			{name: "a empty", a: &empty, b: &foo, expected: false},
			{name: "b empty", a: &foo, b: &empty, expected: false},
			{name: "both equal", a: &foo, b: &foo, expected: true},
			{name: "both different", a: &foo, b: &bar, expected: false},
		}

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				actual := utils.EqualStringPtr(tc.a, tc.b)
				require.Equal(t, tc.expected, actual)
			})
		}

	})

	t.Run("EqualTimePtr", func(t *testing.T) {
		t.Parallel()

		now := time.Now()

		tt := []struct {
			name     string
			a        *time.Time
			b        *time.Time
			expected bool
		}{
			{name: "both nil", a: nil, b: nil, expected: true},
			{name: "a nil", a: nil, b: &time.Time{}, expected: false},
			{name: "b nil", a: &time.Time{}, b: nil, expected: false},
			{name: "both equal", a: &now, b: &now, expected: true},
			{name: "both different", a: &time.Time{}, b: &now, expected: false},
		}

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				actual := utils.EqualTimePtr(tc.a, tc.b)
				require.Equal(t, tc.expected, actual)
			})
		}
	})
}
