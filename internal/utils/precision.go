package utils

import (
	"fmt"
	"log/slog"
	"math"
	"math/big"
)

// ConvertPrecision adjusts the precision of an integer number.
func ConvertPrecision(n string, currentPrecision uint64, targetPrecision uint64) (*big.Int, error) {
	slog.Debug("Current precision", "currentPrecision", currentPrecision)
	slog.Debug("Target precision", "targetPrecision", targetPrecision)

	if currentPrecision < 1 {
		return nil, fmt.Errorf("invalid current precision: %d", currentPrecision)
	}

	if targetPrecision < 1 {
		return nil, fmt.Errorf("invalid target precision: %d", targetPrecision)
	}

	if currentPrecision == targetPrecision {
		return nil, fmt.Errorf("current precision is equal to target precision: %d", currentPrecision)
	}

	// `n` is a string representation of an integer
	// The `SetString` method doesn't support scientific notation even if the number is an integer
	bi := new(big.Int)
	_, ok := bi.SetString(n, 10)
	if !ok {
		return nil, fmt.Errorf("error parsing big.Int: %s", n)
	}

	// Calculate the difference in precision
	precisionDiff := int64(targetPrecision) - int64(currentPrecision)
	slog.Debug("Precision difference", "precisionDiff", precisionDiff)
	absPrecisionDiff := int64(math.Abs(float64(precisionDiff)))
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(absPrecisionDiff), nil)
	slog.Debug("Multiplier", "multiplier", multiplier)

	var result *big.Int
	if precisionDiff > 0 {
		//	// Increase precision by multiplying
		result = new(big.Int).Mul(bi, multiplier)
	} else {
		//	// Decrease precision by dividing
		result = new(big.Int).Div(bi, multiplier)
	}

	slog.Debug("Conversion result", "result", result)

	if result.Cmp(big.NewInt(0)) <= 0 {
		return nil, fmt.Errorf("amount after conversion is less than or equal to 0: %d", result)
	}

	return result, nil
}
