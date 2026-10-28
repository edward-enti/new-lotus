package domain

import (
	"errors"
	"math"
)

// Paint calculator (SPEC §6.1). Pure math only: no I/O, no repository
// dependency. Placed here (PLAN §7 Gate 1 interpretation) so the
// SPEC-named repository test file can test it without importing the
// service package (which would be a layer inversion / import cycle).

// CalculateLiters returns total paint needed:
// Total = Area (m2) x Coats / Coverage Rate x 1.05 (5% wastage).
// Errors: negative area, negative coats, coverage <= 0.
// Zero area or zero coats need no paint and return 0 without error.
func CalculateLiters(areaM2 float64, coats int, coverageRate float64) (float64, error) {
	if areaM2 < 0 {
		return 0, errors.New("calculator: area must not be negative")
	}
	if coats < 0 {
		return 0, errors.New("calculator: coats must not be negative")
	}
	if coverageRate <= 0 {
		return 0, errors.New("calculator: coverage rate must be positive")
	}
	if areaM2 == 0 || coats == 0 {
		return 0, nil
	}
	return areaM2 * float64(coats) / coverageRate * 1.05, nil
}

// PackagingMix is a count of standard pack sizes for a paint volume.
type PackagingMix struct {
	Pack18L int // Thung 18L
	Pack38L int // Lon 3.8L
	Pack08L int // Lon 0.8L
}

// PackagingMixFor returns a greedy pack mix that covers liters with the
// smallest overshoot the greedy pass produces: largest packs first, and
// one smallest pack to cover any fractional remainder. Zero or negative
// input returns an empty mix.
func PackagingMixFor(liters float64) PackagingMix {
	if liters <= 0 {
		return PackagingMix{}
	}
	remaining := liters
	mix := PackagingMix{}
	mix.Pack18L = int(remaining / 18)
	remaining -= float64(mix.Pack18L) * 18
	mix.Pack38L = int(remaining / 3.8)
	remaining -= float64(mix.Pack38L) * 3.8
	mix.Pack08L = int(remaining / 0.8)
	remaining -= float64(mix.Pack08L) * 0.8
	if remaining > 0 {
		// Fractional remainder: one more smallest pack covers it.
		// math.Ceil guard keeps this correct even for tiny remainders
		// caused by floating-point rounding.
		extra := int(math.Ceil(remaining / 0.8))
		mix.Pack08L += extra
	}
	return mix
}
