package utils_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/liftedinit/mfx-migrator/internal/utils"
)

func TestConvertPrecision(t *testing.T) {
	n1 := "1"                                // 1e1
	n3 := "1000"                             // 1e3
	n24 := "1000000000000000000000000"       // 1e24
	n30 := "1000000000000000000000000000000" // 1e30
	n18 := "1000000000000000000"             // 1e18
	n12 := "1000000000000"                   // 1e12
	nF := "12.345"

	bigIntValue30, _ := new(big.Int).SetString(n30, 10)
	bigIntValue12, _ := new(big.Int).SetString(n12, 10)

	tt := []struct {
		name             string
		n                string
		currentPrecision uint64
		targetPrecision  uint64
		expected         *big.Int
		err              string
	}{
		{name: "no precision change", n: n18, currentPrecision: 18, targetPrecision: 18, expected: nil, err: "current precision is equal to target precision: 18"},
		{name: "increase precision", n: n18, currentPrecision: 18, targetPrecision: 24, expected: bigIntValue12},
		{name: "decrease precision", n: n24, currentPrecision: 24, targetPrecision: 18, expected: bigIntValue30},
		{name: "decrease precision 2", n: n3, currentPrecision: 9, targetPrecision: 6, expected: big.NewInt(1000000)},
		{name: "decrease precision 3", n: n1, currentPrecision: 9, targetPrecision: 6, expected: big.NewInt(1000)},
		{name: "invalid conversion (amount <= 0)", n: n1, currentPrecision: 1, targetPrecision: 3, err: "amount after conversion is less than or equal to 0: 0"},
		{name: "invalid number (scientific notation)", n: "1e18", currentPrecision: 18, targetPrecision: 24, err: "error parsing big.Int: 1e18"},
		{name: "invalid number (not a number)", n: "foo", currentPrecision: 18, targetPrecision: 24, err: "error parsing big.Int: foo"},
		{name: "invalid number (empty string)", n: "", currentPrecision: 18, targetPrecision: 24, err: "error parsing big.Int: "},
		{name: "invalid number (fractional)", n: nF, currentPrecision: 3, targetPrecision: 24, err: "error parsing big.Int: 12.345"},
		{name: "invalid current precision", n: n18, currentPrecision: 0, targetPrecision: 24, err: "invalid current precision: 0"},
		{name: "invalid target precision", n: n18, currentPrecision: 18, targetPrecision: 0, err: "invalid target precision: 0"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := utils.ConvertPrecision(tc.n, tc.currentPrecision, tc.targetPrecision)
			if tc.err != "" {
				require.ErrorContains(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}
