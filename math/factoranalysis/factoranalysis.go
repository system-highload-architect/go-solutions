// Package factoranalysis provides Principal Component Analysis (PCA) and related methods
// for dimensionality reduction and exploratory data analysis.
package factoranalysis

import (
	"errors"
	"math"
	// будем использовать наш же statistics
)

// PCA represents a trained Principal Component Analysis model.
type PCA struct {
	Components        [][]float64 // matrix (nComponents × nFeatures)
	ExplainedVariance []float64   // variance explained by each component
	Mean              []float64   // mean of the training data (for centering)
	nComponents       int
}

// TrainPCA fits a PCA model to the given data and returns the first nComponents.
func TrainPCA(X [][]float64, nComponents int) (*PCA, error) {
	if len(X) == 0 || len(X[0]) == 0 {
		return nil, errors.New("factoranalysis: empty dataset")
	}
	if nComponents < 1 || nComponents > len(X[0]) {
		return nil, errors.New("factoranalysis: invalid number of components")
	}

	// 1. Center the data (subtract mean)
	nSamples := len(X)
	nFeatures := len(X[0])
	mean := make([]float64, nFeatures)
	for _, row := range X {
		for j, v := range row {
			mean[j] += v
		}
	}
	for j := range mean {
		mean[j] /= float64(nSamples)
	}
	centered := make([][]float64, nSamples)
	for i, row := range X {
		centered[i] = make([]float64, nFeatures)
		for j, v := range row {
			centered[i][j] = v - mean[j]
		}
	}

	// 2. Compute covariance matrix (nFeatures × nFeatures)
	cov := make([][]float64, nFeatures)
	for i := range cov {
		cov[i] = make([]float64, nFeatures)
	}
	for i := 0; i < nFeatures; i++ {
		for j := i; j < nFeatures; j++ {
			var s float64
			for k := 0; k < nSamples; k++ {
				s += centered[k][i] * centered[k][j]
			}
			cov[i][j] = s / float64(nSamples-1)
			cov[j][i] = cov[i][j]
		}
	}

	// 3. Eigendecomposition using power iteration (simplified for real symmetric matrix)
	eigVals, eigVecs, err := eigen(cov, nComponents)
	if err != nil {
		return nil, err
	}

	// 4. Sort by descending eigenvalue (already done in power iteration)
	return &PCA{
		Components:        eigVecs,
		ExplainedVariance: eigVals,
		Mean:              mean,
		nComponents:       nComponents,
	}, nil
}

// Transform projects the data onto the principal components.
func (p *PCA) Transform(X [][]float64) [][]float64 {
	nSamples := len(X)
	result := make([][]float64, nSamples)
	for i, row := range X {
		result[i] = make([]float64, p.nComponents)
		for c := 0; c < p.nComponents; c++ {
			var sum float64
			for j, v := range row {
				sum += (v - p.Mean[j]) * p.Components[c][j]
			}
			result[i][c] = sum
		}
	}
	return result
}

// InverseTransform reconstructs the original data from the principal components (approximation).
func (p *PCA) InverseTransform(Z [][]float64) [][]float64 {
	nSamples := len(Z)
	nFeatures := len(p.Mean)
	result := make([][]float64, nSamples)
	for i, z := range Z {
		result[i] = make([]float64, nFeatures)
		for j := 0; j < nFeatures; j++ {
			var sum float64
			for c := 0; c < p.nComponents; c++ {
				sum += z[c] * p.Components[c][j]
			}
			result[i][j] = sum + p.Mean[j]
		}
	}
	return result
}

// ExplainedVarianceRatio returns the proportion of variance explained by each component.
func (p *PCA) ExplainedVarianceRatio() []float64 {
	total := 0.0
	for _, v := range p.ExplainedVariance {
		total += v
	}
	if total == 0 {
		return make([]float64, len(p.ExplainedVariance))
	}
	ratios := make([]float64, len(p.ExplainedVariance))
	for i, v := range p.ExplainedVariance {
		ratios[i] = v / total
	}
	return ratios
}

// Varimax rotates the components to maximise the variance of squared loadings (simplified).
// This is a post‑processing step for better interpretability.
func (p *PCA) Varimax(gamma float64, maxIter int) *PCA {
	nComp := p.nComponents
	nFeat := len(p.Mean)
	rot := make([][]float64, nComp)
	for i := range rot {
		rot[i] = make([]float64, nFeat)
		copy(rot[i], p.Components[i])
	}
	for iter := 0; iter < maxIter; iter++ {
		var maxDelta float64
		for i := 0; i < nComp-1; i++ {
			for j := i + 1; j < nComp; j++ {
				// Compute the rotation angle
				var a, b, c, d float64
				for k := 0; k < nFeat; k++ {
					u := rot[i][k]
					v := rot[j][k]
					a += u*u - v*v
					b += 2 * u * v
					c += u*u*u*u - 6*u*u*v*v + v*v*v*v
					d += 4 * u * v * (u*u - v*v)
				}
				num := d - 2*a*b/float64(nFeat)
				den := c - (a*a-b*b)/float64(nFeat)
				phi := 0.0
				if den != 0 {
					phi = 0.25 * math.Atan2(num, den)
				}
				cos := math.Cos(phi)
				sin := math.Sin(phi)
				for k := 0; k < nFeat; k++ {
					u := rot[i][k]
					v := rot[j][k]
					rot[i][k] = u*cos + v*sin
					rot[j][k] = -u*sin + v*cos
				}
				delta := math.Abs(sin) + math.Abs(1-cos)
				if delta > maxDelta {
					maxDelta = delta
				}
			}
		}
		if maxDelta < gamma {
			break
		}
	}
	return &PCA{
		Components:        rot,
		ExplainedVariance: p.ExplainedVariance,
		Mean:              p.Mean,
		nComponents:       p.nComponents,
	}
}

// eigen performs eigendecomposition of a symmetric matrix using power iteration.
// Returns the top k eigenvalues and eigenvectors.
func eigen(A [][]float64, k int) ([]float64, [][]float64, error) {
	n := len(A)
	if n == 0 {
		return nil, nil, errors.New("factoranalysis: empty matrix")
	}
	eigVals := make([]float64, k)
	eigVecs := make([][]float64, k)
	// Power iteration with deflation
	for comp := 0; comp < k; comp++ {
		// Initialise a random vector
		vec := make([]float64, n)
		for i := range vec {
			vec[i] = 1.0 // simple initialisation
		}
		// Iterate until convergence
		for iter := 0; iter < 1000; iter++ {
			newVec := make([]float64, n)
			for i := 0; i < n; i++ {
				var sum float64
				for j := 0; j < n; j++ {
					sum += A[i][j] * vec[j]
				}
				newVec[i] = sum
			}
			// Normalise
			norm := 0.0
			for _, v := range newVec {
				norm += v * v
			}
			norm = math.Sqrt(norm)
			if norm == 0 {
				break
			}
			for i := range newVec {
				newVec[i] /= norm
			}
			// Check convergence
			diff := 0.0
			for i := 0; i < n; i++ {
				diff += math.Abs(newVec[i] - vec[i])
			}
			vec = newVec
			if diff < 1e-10 {
				break
			}
		}
		// Eigenvalue = Rayleigh quotient
		var lambda float64
		for i := 0; i < n; i++ {
			var sum float64
			for j := 0; j < n; j++ {
				sum += A[i][j] * vec[j]
			}
			lambda += vec[i] * sum
		}
		eigVals[comp] = lambda
		eigVecs[comp] = vec
		// Deflate A
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				A[i][j] -= lambda * vec[i] * vec[j]
			}
		}
	}
	return eigVals, eigVecs, nil
}
