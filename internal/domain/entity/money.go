package entity

import (
	"encoding/json"
	"fmt"
	"math"
)

// Money represents a monetary value in the smallest currency unit (e.g., cents, kobo, pence).
// All internal calculations use this integer type to avoid floating-point rounding errors.
// 1 major unit = 100 minor units (e.g., 1 USD = 100 cents, 1 NGN = 100 kobo).
//
// Adapt the comment and helper method names to your project's currency:
//   - FromDollars / ToCents  for USD
//   - FromNaira  / ToKobo   for NGN
//   - FromPounds / ToPence  for GBP
type Money int64

// FromMajor converts from the major currency unit (e.g., dollars) to Money (minor units).
func FromMajor(amount float64) Money {
	return Money(math.Round(amount * 100))
}

// ToMajor converts Money (minor units) to the major unit as float64 for display.
func (m Money) ToMajor() float64 {
	return float64(m) / 100.0
}

// ToMinor returns the raw minor-unit value as int64 — use when sending to payment gateways.
func (m Money) ToMinor() int64 {
	return int64(m)
}

func (m Money) Add(other Money) Money      { return m + other }
func (m Money) Subtract(other Money) Money { return m - other }
func (m Money) Multiply(n int) Money       { return m * Money(n) }
func (m Money) IsPositive() bool           { return m > 0 }
func (m Money) IsNegative() bool           { return m < 0 }
func (m Money) IsZero() bool               { return m == 0 }

// String formats Money for logging. Update the symbol to match your currency.
func (m Money) String() string {
	return fmt.Sprintf("$%.2f", m.ToMajor())
}

// MarshalJSON serializes Money as a float (major unit) for API responses.
func (m Money) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.ToMajor())
}

// UnmarshalJSON deserializes a float (major unit) from API requests into Money.
func (m *Money) UnmarshalJSON(data []byte) error {
	var f float64
	if err := json.Unmarshal(data, &f); err != nil {
		return err
	}
	*m = FromMajor(f)
	return nil
}
