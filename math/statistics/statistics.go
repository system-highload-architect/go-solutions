// Package statistics provides descriptive and inferential statistical functions
// optimised for high performance and zero allocations on the hot path.
package statistics

import (
	"math"
	"sort"
)

// Mean returns the arithmetic mean of a slice of float64 values.
func Mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// Variance returns the population variance.
func Variance(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	mean := Mean(values)
	var s2 float64
	for _, v := range values {
		d := v - mean
		s2 += d * d
	}
	return s2 / float64(len(values))
}

// StdDev returns the population standard deviation.
func StdDev(values []float64) float64 { return math.Sqrt(Variance(values)) }

// Covariance computes the covariance between two equal-length slices.
func Covariance(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0
	}
	mx, my := Mean(x), Mean(y)
	var cov float64
	for i := range x {
		cov += (x[i] - mx) * (y[i] - my)
	}
	return cov / float64(len(x)-1) // sample covariance by default
}

// Correlation computes the Pearson correlation coefficient.
func Correlation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0
	}
	sx, sy := StdDev(x), StdDev(y)
	if sx == 0 || sy == 0 {
		return 0
	}
	return Covariance(x, y) / (sx * sy)
}

// Percentile computes the p-th percentile using linear interpolation.
// p must be between 0 and 1.
func Percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	if p <= 0 {
		return values[0]
	}
	if p >= 1 {
		return values[len(values)-1]
	}
	// copy and sort
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	idx := p * float64(len(values)-1)
	lo := int(math.Floor(idx))
	hi := int(math.Ceil(idx))
	if lo == hi {
		return sorted[lo]
	}
	return sorted[lo] + (sorted[hi]-sorted[lo])*(idx-float64(lo))
}

// Median returns the 50th percentile.
func Median(values []float64) float64 { return Percentile(values, 0.5) }

// MinMax returns the minimum and maximum values.
func MinMax(values []float64) (min, max float64) {
	if len(values) == 0 {
		return 0, 0
	}
	min, max = values[0], values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return
}

// Sum returns the sum of all values.
func Sum(values []float64) float64 {
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum
}

// Skewness computes the sample skewness.
func Skewness(values []float64) float64 {
	n := len(values)
	if n < 3 {
		return 0
	}
	mean := Mean(values)
	var m2, m3 float64
	for _, v := range values {
		d := v - mean
		m2 += d * d
		m3 += d * d * d
	}
	return (float64(n) * m3) / ((float64(n) - 1) * (float64(n) - 2) * math.Pow(m2/float64(n), 1.5))
}

// Kurtosis computes the sample excess kurtosis.
func Kurtosis(values []float64) float64 {
	n := len(values)
	if n < 4 {
		return 0
	}
	mean := Mean(values)
	var m2, m4 float64
	for _, v := range values {
		d := v - mean
		m2 += d * d
		m4 += d * d * d * d
	}
	m2 /= float64(n)
	m4 /= float64(n)
	return (m4/(m2*m2) - 3) * (float64(n) - 1) * (float64(n) + 1) / ((float64(n) - 2) * (float64(n) - 3))
}

// Entropy computes the Shannon entropy of a probability distribution.
func Entropy(probs []float64) float64 {
	var entropy float64
	for _, p := range probs {
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}
	return entropy
}

// Normalise applies z-score normalisation and returns a new slice.
func Normalise(values []float64) []float64 {
	if len(values) < 2 {
		return values
	}
	mean, sd := Mean(values), StdDev(values)
	if sd == 0 {
		return make([]float64, len(values))
	}
	out := make([]float64, len(values))
	for i, v := range values {
		out[i] = (v - mean) / sd
	}
	return out
}

// Quantile returns the value at a given quantile using the R-8 method.
func Quantile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	n := float64(len(sorted))
	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[len(sorted)-1]
	}
	idx := p*(n+1.0/3.0) - 1.0/3.0
	lo := int(math.Floor(idx))
	hi := int(math.Ceil(idx))
	if lo == hi {
		return sorted[lo]
	}
	frac := idx - float64(lo)
	return sorted[lo] + frac*(sorted[hi]-sorted[lo])
}
