// Package valuation provides models for estimating the value of advertising
// impressions and optimising bid prices in Real‑Time Bidding systems.
package valuation

import (
	"math"

	"github.com/system-highload-architect/go-solutions/geo/geospatial"
	"github.com/system-highload-architect/go-solutions/math/fixedpoint"
)

// ----------------------------------------------------------------------------
// LTV Model
// ----------------------------------------------------------------------------

// LTVModel predicts the lifetime value of a user based on a feature vector.
type LTVModel struct {
	Coefficients []float64 // Coefficients[0] is the intercept
}

// NewLTVModel creates a new LTVModel with the given coefficients.
func NewLTVModel(coeffs []float64) *LTVModel {
	return &LTVModel{Coefficients: coeffs}
}

// Predict computes the LTV for a given feature vector.
func (m *LTVModel) Predict(features []float64) float64 {
	if len(features) == 0 || len(m.Coefficients) == 0 {
		return 0
	}
	if len(m.Coefficients) == len(features) {
		var sum float64
		for i, v := range features {
			sum += m.Coefficients[i] * v
		}
		return sum
	}
	if len(m.Coefficients) == len(features)+1 {
		result := m.Coefficients[0]
		for i, v := range features {
			result += m.Coefficients[i+1] * v
		}
		return result
	}
	return 0
}

// ----------------------------------------------------------------------------
// Impression Value
// ----------------------------------------------------------------------------

// ImpressionValue estimates the monetary value of a single ad impression.
type ImpressionValue struct {
	BaseCTR             float64
	BaseCVR             float64
	BaseConversionValue float64 // value of one conversion (in monetary units)
}

// NewImpressionValue creates a new ImpressionValue estimator.
func NewImpressionValue(ctr, cvr, conversionValue float64) *ImpressionValue {
	return &ImpressionValue{
		BaseCTR:             ctr,
		BaseCVR:             cvr,
		BaseConversionValue: conversionValue,
	}
}

// Value returns the estimated value of an impression for the given user and ad features.
func (iv *ImpressionValue) Value(userFeatures, adFeatures []float64) float64 {
	return iv.BaseCTR * iv.BaseCVR * iv.BaseConversionValue
}

// ----------------------------------------------------------------------------
// Geo Factor
// ----------------------------------------------------------------------------

// GeoFactor adjusts the value of an impression based on distance.
type GeoFactor struct {
	DecayRate float64 // how quickly the factor decreases with distance (metres)
}

// NewGeoFactor creates a new GeoFactor with the given decay rate.
func NewGeoFactor(decayRate float64) *GeoFactor {
	return &GeoFactor{DecayRate: decayRate}
}

// Factor computes the geo factor for a given distance.
func (gf *GeoFactor) Factor(userPos, targetPos geospatial.Point) float64 {
	dist := geospatial.HaversineDistance(userPos, targetPos)
	return math.Exp(-dist / gf.DecayRate)
}

// ----------------------------------------------------------------------------
// Win Rate Model
// ----------------------------------------------------------------------------

// WinRateModel estimates the probability of winning an RTB auction given a bid price.
type WinRateModel struct {
	Beta0 float64
	Beta1 float64
}

// NewWinRateModel creates a new WinRateModel.
func NewWinRateModel(beta0, beta1 float64) *WinRateModel {
	return &WinRateModel{Beta0: beta0, Beta1: beta1}
}

// Probability returns the estimated win probability for a given bid price.
func (m *WinRateModel) Probability(price fixedpoint.Money) float64 {
	p := float64(price.Amount())
	return 1.0 / (1.0 + math.Exp(-(m.Beta0 + m.Beta1*p)))
}

// OptimalBid computes the bid that maximises the expected profit.
func (m *WinRateModel) OptimalBid(value fixedpoint.Money) (fixedpoint.Money, error) {
	bestBid := fixedpoint.New(0, value.Scale())
	bestProfit := 0.0
	step := int64(1000)
	for bid := int64(0); bid <= value.Amount(); bid += step {
		bidMoney := fixedpoint.New(bid, value.Scale())
		p := m.Probability(bidMoney)
		profit := (float64(value.Amount()) - float64(bid)) * p
		if profit > bestProfit {
			bestProfit = profit
			bestBid = bidMoney
		}
	}
	return bestBid, nil
}

// ----------------------------------------------------------------------------
// Composite Scorer
// ----------------------------------------------------------------------------

// Scorer combines LTV, impression value, geo factor, and win‑rate model.
type Scorer struct {
	LTVModel      *LTVModel
	ImpressionVal *ImpressionValue
	GeoFactor     *GeoFactor
	WinRateModel  *WinRateModel
}

// NewScorer creates a new composite Scorer.
func NewScorer(ltv *LTVModel, iv *ImpressionValue, gf *GeoFactor, wrm *WinRateModel) *Scorer {
	return &Scorer{
		LTVModel:      ltv,
		ImpressionVal: iv,
		GeoFactor:     gf,
		WinRateModel:  wrm,
	}
}

// Score computes the overall score and optimal bid for a given user and campaign.
func (s *Scorer) Score(
	userPos, targetPos geospatial.Point,
	features []float64,
	baseBid fixedpoint.Money,
) (score float64, optimalBid fixedpoint.Money, err error) {
	ltv := s.LTVModel.Predict(features)
	impVal := s.ImpressionVal.Value(features, nil)
	gf := s.GeoFactor.Factor(userPos, targetPos)

	score = ltv * impVal * gf

	if s.WinRateModel != nil {
		optimalBid, err = s.WinRateModel.OptimalBid(baseBid)
	} else {
		optimalBid = baseBid
	}
	return
}
