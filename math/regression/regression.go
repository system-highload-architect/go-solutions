// Package regression provides linear and logistic regression models.
// All prediction methods are allocation‑free.
package regression

import (
	"errors"
	"math"
)

// LinearModel represents a trained linear regression model.
// Coefficients[0] is the intercept, Coefficients[1:] are the slopes.
type LinearModel struct {
	Coefficients []float64
}

// Predict computes y = w0 + w1*x1 + w2*x2 + ...
func (m *LinearModel) Predict(features []float64) float64 {
	if len(features) != len(m.Coefficients)-1 {
		return 0 // or panic? return 0 for safety
	}
	result := m.Coefficients[0]
	for i, v := range features {
		result += m.Coefficients[i+1] * v
	}
	return result
}

// TrainLinear fits a linear regression model using Ordinary Least Squares.
// X is a matrix where each row is a sample, each column is a feature.
// y is the target values.
func TrainLinear(X [][]float64, y []float64) (*LinearModel, error) {
	if len(X) == 0 || len(X) != len(y) {
		return nil, errors.New("regression: empty or mismatched data")
	}
	nSamples := len(X)
	nFeatures := len(X[0])
	if nFeatures == 0 {
		return nil, errors.New("regression: no features")
	}
	// Add intercept column (ones) to X
	augX := make([][]float64, nSamples)
	for i := range augX {
		augX[i] = make([]float64, nFeatures+1)
		augX[i][0] = 1.0 // intercept
		copy(augX[i][1:], X[i])
	}
	// Compute (X^T * X)^-1 * X^T * y
	XTX := make([][]float64, nFeatures+1)
	for i := range XTX {
		XTX[i] = make([]float64, nFeatures+1)
	}
	for i := 0; i < nFeatures+1; i++ {
		for j := 0; j < nFeatures+1; j++ {
			var sum float64
			for k := 0; k < nSamples; k++ {
				sum += augX[k][i] * augX[k][j]
			}
			XTX[i][j] = sum
		}
	}
	XTY := make([]float64, nFeatures+1)
	for i := 0; i < nFeatures+1; i++ {
		var sum float64
		for k := 0; k < nSamples; k++ {
			sum += augX[k][i] * y[k]
		}
		XTY[i] = sum
	}
	// Invert XTX (Gauss‑Jordan)
	inv, err := invertMatrix(XTX)
	if err != nil {
		return nil, err
	}
	// Multiply inv * XTY
	coeffs := make([]float64, nFeatures+1)
	for i := 0; i < nFeatures+1; i++ {
		var sum float64
		for j := 0; j < nFeatures+1; j++ {
			sum += inv[i][j] * XTY[j]
		}
		coeffs[i] = sum
	}
	return &LinearModel{Coefficients: coeffs}, nil
}

// invertMatrix inverts a square matrix using Gauss‑Jordan elimination.
func invertMatrix(A [][]float64) ([][]float64, error) {
	n := len(A)
	// Augment with identity
	aug := make([][]float64, n)
	for i := range aug {
		aug[i] = make([]float64, 2*n)
		copy(aug[i][:n], A[i])
		aug[i][n+i] = 1.0
	}
	for i := 0; i < n; i++ {
		// Pivot
		pivot := aug[i][i]
		if math.Abs(pivot) < 1e-12 {
			return nil, errors.New("regression: singular matrix")
		}
		for j := 0; j < 2*n; j++ {
			aug[i][j] /= pivot
		}
		for k := 0; k < n; k++ {
			if k != i {
				factor := aug[k][i]
				for j := 0; j < 2*n; j++ {
					aug[k][j] -= factor * aug[i][j]
				}
			}
		}
	}
	inv := make([][]float64, n)
	for i := range inv {
		inv[i] = make([]float64, n)
		copy(inv[i], aug[i][n:])
	}
	return inv, nil
}

// LogisticModel represents a trained logistic regression model.
type LogisticModel struct {
	Coefficients []float64
}

// PredictProb computes the probability using the logistic function.
func (m *LogisticModel) PredictProb(features []float64) float64 {
	if len(features) != len(m.Coefficients)-1 {
		return 0
	}
	z := m.Coefficients[0]
	for i, v := range features {
		z += m.Coefficients[i+1] * v
	}
	return 1.0 / (1.0 + math.Exp(-z))
}

// PredictClass returns the binary class (0 or 1) based on a threshold (default 0.5).
func (m *LogisticModel) PredictClass(features []float64, threshold float64) int {
	if m.PredictProb(features) >= threshold {
		return 1
	}
	return 0
}

// TrainLogistic fits a logistic regression model using gradient descent.
// X: feature matrix (samples × features), y: binary labels (0/1),
// learningRate: step size, iterations: number of gradient steps.
func TrainLogistic(X [][]float64, y []float64, learningRate float64, iterations int) (*LogisticModel, error) {
	if len(X) == 0 || len(X) != len(y) {
		return nil, errors.New("regression: empty or mismatched data")
	}
	nFeatures := len(X[0])
	coeffs := make([]float64, nFeatures+1) // + intercept
	for iter := 0; iter < iterations; iter++ {
		grad := make([]float64, len(coeffs))
		for i, features := range X {
			pred := coeffs[0]
			for j, v := range features {
				pred += coeffs[j+1] * v
			}
			pred = 1.0 / (1.0 + math.Exp(-pred)) // sigmoid
			err := pred - y[i]
			grad[0] += err
			for j, v := range features {
				grad[j+1] += err * v
			}
		}
		n := float64(len(X))
		for j := range coeffs {
			coeffs[j] -= learningRate * grad[j] / n
		}
	}
	return &LogisticModel{Coefficients: coeffs}, nil
}
