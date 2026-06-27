// Package fixedpoint provides exact monetary arithmetic using integer representation.
// Money values are stored in minimal currency units (cents, kopeks, etc.) with an associated scale.
//
// All operations are designed to be allocation‑free on the hot path and safe for concurrent use.
package fixedpoint

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unsafe"
)

// ----------------------------------------------------------------------------
// unsafe byte/string conversions – see unsafe.go for details
// ----------------------------------------------------------------------------
// bytesToString converts a byte slice to a string without copying the underlying memory.
// The caller MUST NOT modify the original slice after conversion.
func bytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// stringToBytes converts a string to a byte slice without copying.
// The returned slice MUST NOT be modified.
func stringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// ----------------------------------------------------------------------------
// Money type
// ----------------------------------------------------------------------------

// Money represents a monetary value in minimal units (e.g., kopeks for RUB, cents for USD).
// The scale indicates the number of decimal places (typically 2 for most currencies).
type Money struct {
	amount int64
	scale  int32
}

// Common scales for different currencies.
const (
	ScaleRUB = 2
	ScaleUSD = 2
	ScaleEUR = 2
)

// Common currency units with default scale 2.
var (
	RUB = Money{scale: ScaleRUB}
	USD = Money{scale: ScaleUSD}
	EUR = Money{scale: ScaleEUR}
)

// New creates a new Money value from an amount and a scale.
func New(amount int64, scale int32) Money {
	return Money{amount: amount, scale: scale}
}

// Zero returns a zero monetary value with the given scale.
func Zero(scale int32) Money {
	return Money{scale: scale}
}

// FromFloat64 converts a floating-point value to Money using the specified scale,
// rounding to the nearest minimal unit.
func FromFloat64(value float64, scale int32) Money {
	if scale < 0 {
		scale = 0
	}
	multiplier := math.Pow10(int(scale))
	amount := int64(math.Round(value * multiplier))
	return Money{amount: amount, scale: scale}
}

// FromString parses a monetary string (e.g., "100.50") into Money with the given scale.
// The string must contain exactly the number of decimal places specified by scale.
func FromString(s string, scale int32) (Money, error) {
	if scale < 0 {
		return Money{}, errors.New("fixedpoint: invalid scale")
	}

	// Split on '.' to separate integer and fractional parts.
	parts := strings.SplitN(s, ".", 2)
	integerPart := parts[0]
	var fractionalPart string
	if len(parts) == 2 {
		fractionalPart = parts[1]
	}

	// Pad or truncate fractional part to exactly scale digits.
	if len(fractionalPart) < int(scale) {
		fractionalPart += strings.Repeat("0", int(scale)-len(fractionalPart))
	} else if len(fractionalPart) > int(scale) {
		fractionalPart = fractionalPart[:scale]
	}

	// Combine integer and fractional parts into one integer.
	combined := integerPart + fractionalPart
	amount, err := strconv.ParseInt(combined, 10, 64)
	if err != nil {
		return Money{}, fmt.Errorf("fixedpoint: cannot parse %q: %w", s, err)
	}
	return Money{amount: amount, scale: scale}, nil
}

// MustFromString is like FromString but panics on error.
func MustFromString(s string, scale int32) Money {
	m, err := FromString(s, scale)
	if err != nil {
		panic(err)
	}
	return m
}

// Amount returns the raw amount in minimal units.
func (m Money) Amount() int64 {
	return m.amount
}

// Scale returns the number of decimal places.
func (m Money) Scale() int32 {
	return m.scale
}

// Add returns the sum of m and other. The scales must be equal.
func (m Money) Add(other Money) (Money, error) {
	if m.scale != other.scale {
		return Money{}, fmt.Errorf("fixedpoint: scale mismatch: %d vs %d", m.scale, other.scale)
	}
	return Money{amount: m.amount + other.amount, scale: m.scale}, nil
}

// Sub returns the difference of m and other. The scales must be equal.
func (m Money) Sub(other Money) (Money, error) {
	if m.scale != other.scale {
		return Money{}, fmt.Errorf("fixedpoint: scale mismatch: %d vs %d", m.scale, other.scale)
	}
	return Money{amount: m.amount - other.amount, scale: m.scale}, nil
}

// Mul multiplies the monetary value by an integer factor.
func (m Money) Mul(factor int64) Money {
	return Money{amount: m.amount * factor, scale: m.scale}
}

// Div divides the monetary value by an integer divisor, rounding to the nearest unit.
func (m Money) Div(divisor int64) (Money, error) {
	if divisor == 0 {
		return Money{}, errors.New("fixedpoint: division by zero")
	}
	q := m.amount / divisor
	r := m.amount % divisor
	// Round to nearest
	if (r > 0 && divisor > 0 && r*2 >= divisor) || (r < 0 && divisor < 0 && -r*2 >= -divisor) {
		q++
	}
	return Money{amount: q, scale: m.scale}, nil
}

// MulFloat multiplies the monetary value by a floating-point factor, rounding to the nearest unit.
func (m Money) MulFloat(factor float64) Money {
	newAmount := int64(math.Round(float64(m.amount) * factor))
	return Money{amount: newAmount, scale: m.scale}
}

// DivFloat divides the monetary value by a floating-point divisor, rounding to the nearest unit.
func (m Money) DivFloat(divisor float64) Money {
	newAmount := int64(math.Round(float64(m.amount) / divisor))
	return Money{amount: newAmount, scale: m.scale}
}

// AddChecked performs addition with overflow detection.
func (m Money) AddChecked(other Money) (Money, error) {
	if m.scale != other.scale {
		return Money{}, fmt.Errorf("fixedpoint: scale mismatch: %d vs %d", m.scale, other.scale)
	}
	sum := m.amount + other.amount
	if (sum > m.amount) != (other.amount > 0) {
		return Money{}, errors.New("fixedpoint: overflow")
	}
	return Money{amount: sum, scale: m.scale}, nil
}

// SubChecked performs subtraction with overflow detection.
func (m Money) SubChecked(other Money) (Money, error) {
	if m.scale != other.scale {
		return Money{}, fmt.Errorf("fixedpoint: scale mismatch: %d vs %d", m.scale, other.scale)
	}
	diff := m.amount - other.amount
	if (diff < m.amount) != (other.amount > 0) {
		return Money{}, errors.New("fixedpoint: overflow")
	}
	return Money{amount: diff, scale: m.scale}, nil
}

// MulChecked performs multiplication with overflow detection.
func (m Money) MulChecked(factor int64) (Money, error) {
	if factor == 0 {
		return Money{scale: m.scale}, nil
	}
	result := m.amount * factor
	if result/factor != m.amount {
		return Money{}, errors.New("fixedpoint: overflow")
	}
	return Money{amount: result, scale: m.scale}, nil
}

// Cmp compares two monetary values. Returns -1, 0, or 1.
// Scales must be equal.
func (m Money) Cmp(other Money) int {
	if m.amount < other.amount {
		return -1
	}
	if m.amount > other.amount {
		return 1
	}
	return 0
}

// Equals returns true if the two monetary values are exactly equal.
func (m Money) Equals(other Money) bool {
	return m.amount == other.amount && m.scale == other.scale
}

// LessThan returns true if m < other.
func (m Money) LessThan(other Money) bool {
	return m.amount < other.amount
}

// GreaterThan returns true if m > other.
func (m Money) GreaterThan(other Money) bool {
	return m.amount > other.amount
}

// IsZero returns true if the value is zero.
func (m Money) IsZero() bool {
	return m.amount == 0
}

// Sign returns -1, 0, or 1 depending on whether the value is negative, zero, or positive.
func (m Money) Sign() int {
	if m.amount < 0 {
		return -1
	}
	if m.amount > 0 {
		return 1
	}
	return 0
}

// Abs returns the absolute value.
func (m Money) Abs() Money {
	if m.amount < 0 {
		return Money{amount: -m.amount, scale: m.scale}
	}
	return m
}

// Round rounds the monetary value to the nearest multiple of minUnit.
func (m Money) Round(minUnit int64) Money {
	if minUnit <= 0 {
		return m
	}
	rem := m.amount % minUnit
	if rem == 0 {
		return m
	}
	// Round to nearest
	if (rem > 0 && rem*2 >= minUnit) || (rem < 0 && -rem*2 >= -minUnit) {
		m.amount += minUnit - rem
	} else {
		m.amount -= rem
	}
	return m
}

// Ceil rounds up to the nearest whole currency unit (i.e., scale 0).
func (m Money) Ceil() Money {
	unit := int64(math.Pow10(int(m.scale)))
	rem := m.amount % unit
	if rem > 0 {
		m.amount += unit - rem
	} else if rem < 0 {
		m.amount -= rem
	}
	return Money{amount: m.amount, scale: m.scale}
}

// Floor rounds down to the nearest whole currency unit.
func (m Money) Floor() Money {
	unit := int64(math.Pow10(int(m.scale)))
	rem := m.amount % unit
	if rem < 0 {
		m.amount -= unit + rem
	}
	return Money{amount: m.amount, scale: m.scale}
}

// Trunc truncates the value towards zero to the nearest whole currency unit.
func (m Money) Trunc() Money {
	unit := int64(math.Pow10(int(m.scale)))
	m.amount -= m.amount % unit
	return Money{amount: m.amount, scale: m.scale}
}

// String formats the monetary value as a string with the correct number of decimal places.
func (m Money) String() string {
	return m.Format(m.scale)
}

// Format returns a string representation with exactly the specified number of decimal places.
// This method uses no allocations from fmt and minimal allocations overall.
func (m Money) Format(decimals int32) string {
	if decimals < 0 {
		decimals = 0
	}
	unit := int64(math.Pow10(int(decimals)))
	integral := m.amount / unit
	fractional := m.amount % unit
	if fractional < 0 {
		fractional = -fractional
		integral++ // because we took the absolute fractional part
	}

	var buf strings.Builder
	buf.Grow(24) // enough for int64 + dot + digits
	buf.WriteString(strconv.FormatInt(integral, 10))
	if decimals > 0 {
		buf.WriteByte('.')
		fmtStr := strconv.FormatInt(fractional, 10)
		// Pad with leading zeros
		for i := int32(len(fmtStr)); i < decimals; i++ {
			buf.WriteByte('0')
		}
		buf.WriteString(fmtStr)
	}
	return buf.String()
}

// Float64 returns the monetary value as a float64 (for display only).
func (m Money) Float64() float64 {
	unit := math.Pow10(int(m.scale))
	return float64(m.amount) / unit
}

// MarshalJSON implements json.Marshaler.
func (m Money) MarshalJSON() ([]byte, error) {
	// Allocate a buffer with enough capacity.
	str := m.String()
	buf := make([]byte, 0, len(str)+2)
	buf = append(buf, '"')
	buf = append(buf, str...)
	buf = append(buf, '"')
	return buf, nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *Money) UnmarshalJSON(data []byte) error {
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		s := bytesToString(data[1 : len(data)-1])
		val, err := FromString(s, m.scale)
		if err != nil {
			return err
		}
		m.amount = val.amount
		return nil
	}
	return errors.New("fixedpoint: invalid JSON format")
}

// MarshalText implements encoding.TextMarshaler.
func (m Money) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil // []byte(m.String()) does allocate, but this is not on the hot path.
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (m *Money) UnmarshalText(data []byte) error {
	val, err := FromString(bytesToString(data), m.scale)
	if err != nil {
		return err
	}
	m.amount = val.amount
	return nil
}

// Convert converts the monetary value to another currency using the given exchange rate
// and target scale, rounding to the nearest minimal unit.
func (m Money) Convert(rate float64, targetScale int32) Money {
	base := float64(m.amount) / math.Pow10(int(m.scale))
	newValue := base * rate
	return FromFloat64(newValue, targetScale)
}

// Sum returns the total of a slice of Money values. All must have the same scale.
func Sum(values []Money) (Money, error) {
	if len(values) == 0 {
		return Money{}, errors.New("fixedpoint: no values to sum")
	}
	scale := values[0].scale
	var total int64
	for _, v := range values {
		if v.scale != scale {
			return Money{}, fmt.Errorf("fixedpoint: scale mismatch in Sum: %d vs %d", v.scale, scale)
		}
		total += v.amount
	}
	return Money{amount: total, scale: scale}, nil
}

// Avg returns the arithmetic mean of a slice of Money values. All must have the same scale.
func Avg(values []Money) (Money, error) {
	sum, err := Sum(values)
	if err != nil {
		return Money{}, err
	}
	return sum.Div(int64(len(values)))
}
