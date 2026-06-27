// Package timeseries provides functions for time series analysis and forecasting.
// All implementations are pure Go and optimized for minimal allocations on the hot path.
package timeseries

import (
	"errors"
)

// HoltWintersParams holds the smoothing parameters for Holt‑Winters forecasting.
type HoltWintersParams struct {
	Alpha  float64 // level smoothing (0–1)
	Beta   float64 // trend smoothing (0–1)
	Gamma  float64 // seasonal smoothing (0–1)
	Period int     // length of a seasonal cycle (e.g., 12 for monthly data with yearly seasonality)
}

// HoltWintersForecast performs triple exponential smoothing (Holt‑Winters)
// and returns the forecast for the next `horizon` points.
// If params.Period <= 0, a non‑seasonal model (Holt) is used.
func HoltWintersForecast(data []float64, horizon int, params HoltWintersParams) ([]float64, error) {
	n := len(data)
	if n < 2 {
		return nil, errors.New("timeseries: need at least two observations")
	}
	if horizon < 0 {
		return nil, errors.New("timeseries: horizon must be non-negative")
	}
	if params.Alpha == 0 && params.Beta == 0 && params.Gamma == 0 {
		// sensible defaults
		params.Alpha = 0.5
		params.Beta = 0.3
		params.Gamma = 0.2
		params.Period = 4
	}

	useSeasonal := params.Period > 0 && params.Gamma > 0
	if useSeasonal && n < 2*params.Period {
		// fall back to non‑seasonal (Holt)
		useSeasonal = false
	}

	// --- initialise components ---
	level := make([]float64, n)
	trend := make([]float64, n)
	var seasonal []float64
	if useSeasonal {
		seasonal = make([]float64, n)
		// initial seasonal indices from the first season
		firstSeasonSum := 0.0
		for i := 0; i < params.Period; i++ {
			firstSeasonSum += data[i]
		}
		firstSeasonMean := firstSeasonSum / float64(params.Period)
		for i := 0; i < params.Period; i++ {
			seasonal[i] = data[i] / firstSeasonMean
		}
		level[params.Period-1] = firstSeasonMean
		trend[params.Period-1] = 0
	} else {
		// simple initialisation for Holt / Brown
		level[0] = data[0]
		trend[0] = data[1] - data[0]
	}

	// --- forward pass (filtering) ---
	for t := 1; t < n; t++ {
		if useSeasonal {
			if t >= params.Period {
				level[t] = params.Alpha*(data[t]-seasonal[t-params.Period]) + (1-params.Alpha)*(level[t-1]+trend[t-1])
				trend[t] = params.Beta*(level[t]-level[t-1]) + (1-params.Beta)*trend[t-1]
				seasonal[t] = params.Gamma*(data[t]-level[t]) + (1-params.Gamma)*seasonal[t-params.Period]
			} else {
				// not enough data for seasonality yet; keep previous
				level[t] = level[t-1]
				trend[t] = trend[t-1]
				seasonal[t] = seasonal[t] // initial seasonal constant
			}
		} else {
			// Holt (no seasonality)
			level[t] = params.Alpha*data[t] + (1-params.Alpha)*(level[t-1]+trend[t-1])
			trend[t] = params.Beta*(level[t]-level[t-1]) + (1-params.Beta)*trend[t-1]
		}
	}

	// --- forecast ---
	forecast := make([]float64, horizon)
	for h := 0; h < horizon; h++ {
		t := n - 1
		if useSeasonal {
			seasonIdx := (t + h + 1) % params.Period
			// if we are forecasting beyond one season, wrap around
			seasonalVal := seasonal[0] // safe fallback
			if seasonIdx < n {
				seasonalVal = seasonal[seasonIdx]
			} else {
				// extrapolate by repeating the last season
				seasonalVal = seasonal[t-params.Period+1 : t+1][h%params.Period]
			}
			forecast[h] = (level[t] + float64(h+1)*trend[t]) * seasonalVal
		} else {
			forecast[h] = level[t] + float64(h+1)*trend[t]
		}
	}
	return forecast, nil
}

// SimpleMovingAverage returns the simple moving average (SMA) of a series with a given window.
// The result slice has length len(data)-window+1. If window > len(data), returns nil.
func SimpleMovingAverage(data []float64, window int) []float64 {
	if window <= 0 || window > len(data) {
		return nil
	}
	n := len(data) - window + 1
	result := make([]float64, n)
	var sum float64
	for i := 0; i < window; i++ {
		sum += data[i]
	}
	result[0] = sum / float64(window)
	for i := 1; i < n; i++ {
		sum += data[i+window-1] - data[i-1]
		result[i] = sum / float64(window)
	}
	return result
}

// ExponentialMovingAverage computes the exponentially weighted moving average.
// alpha is the smoothing factor (0 < alpha < 1).
func ExponentialMovingAverage(data []float64, alpha float64) []float64 {
	if len(data) == 0 {
		return nil
	}
	result := make([]float64, len(data))
	result[0] = data[0]
	for i := 1; i < len(data); i++ {
		result[i] = alpha*data[i] + (1-alpha)*result[i-1]
	}
	return result
}

// LinearDetrend removes a linear trend from the series.
// Returns the detrended series and the slope+intercept coefficients.
func LinearDetrend(data []float64) (detrended []float64, intercept, slope float64) {
	n := float64(len(data))
	if n < 2 {
		return data, 0, 0
	}
	// simple linear regression on index
	var sumX, sumY, sumXY, sumX2 float64
	for i, y := range data {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	slope = (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept = (sumY - slope*sumX) / n

	detrended = make([]float64, len(data))
	for i, y := range data {
		detrended[i] = y - (slope*float64(i) + intercept)
	}
	return
}

// RemoveSeasonality subtracts a seasonal component computed by averaging over each period.
// The period must be provided. It returns the deseasonalized series and the seasonal indices.
func RemoveSeasonality(data []float64, period int) (deseasonalized []float64, seasonal []float64) {
	n := len(data)
	if period <= 0 || n < period {
		return data, nil
	}
	seasonal = make([]float64, period)
	counts := make([]float64, period)
	for i, v := range data {
		idx := i % period
		seasonal[idx] += v
		counts[idx]++
	}
	for i := range seasonal {
		if counts[i] > 0 {
			seasonal[i] /= counts[i]
		}
	}
	deseasonalized = make([]float64, n)
	for i, v := range data {
		deseasonalized[i] = v - seasonal[i%period]
	}
	return
}

// AutoCorrelation computes the autocorrelation function up to maxLag.
func AutoCorrelation(data []float64, maxLag int) []float64 {
	n := len(data)
	if n == 0 || maxLag < 0 {
		return nil
	}
	mean := 0.0
	for _, v := range data {
		mean += v
	}
	mean /= float64(n)
	var denom float64
	for _, v := range data {
		denom += (v - mean) * (v - mean)
	}
	if denom == 0 {
		return make([]float64, maxLag+1)
	}
	acf := make([]float64, maxLag+1)
	for lag := 0; lag <= maxLag; lag++ {
		var num float64
		for i := 0; i < n-lag; i++ {
			num += (data[i] - mean) * (data[i+lag] - mean)
		}
		acf[lag] = num / denom
	}
	return acf
}
