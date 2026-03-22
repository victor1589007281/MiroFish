package prediction

import "math"

// Calibrator applies statistical calibration to raw LLM probabilities.
type Calibrator struct {
	A                   float64 // Platt scaling slope (default -1.0, train from historical data)
	B                   float64 // Platt scaling intercept (default 0.0)
	ExtremizationFactor float64 // >1 pushes away from 50% (default 1.5)
}

func NewCalibrator() *Calibrator {
	return &Calibrator{
		A:                   -1.0,
		B:                   0.0,
		ExtremizationFactor: 1.5,
	}
}

// Calibrate applies Platt scaling followed by extremization.
func (c *Calibrator) Calibrate(rawProb float64) float64 {
	rawProb = clamp(rawProb, 0.01, 0.99)

	// Platt scaling: correct systematic LLM bias
	logit := math.Log(rawProb / (1.0 - rawProb))
	plattProb := 1.0 / (1.0 + math.Exp(c.A*logit+c.B))

	// Extremization: LLMs are naturally conservative (biased toward 50%)
	extremized := extremize(plattProb, c.ExtremizationFactor)

	return clamp(extremized, 0.01, 0.99)
}

func extremize(p float64, factor float64) float64 {
	p = clamp(p, 0.001, 0.999)
	logOdds := math.Log(p / (1.0 - p))
	extremeLogOdds := logOdds * factor
	return 1.0 / (1.0 + math.Exp(-extremeLogOdds))
}

func clamp(v, min, max float64) float64 {
	return math.Max(min, math.Min(max, v))
}

// StdDev calculates standard deviation of a slice.
func StdDev(vals []float64) float64 {
	if len(vals) < 2 {
		return 0
	}
	m := Mean(vals)
	var ss float64
	for _, v := range vals {
		d := v - m
		ss += d * d
	}
	return math.Sqrt(ss / float64(len(vals)))
}

// Mean calculates the average.
func Mean(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var s float64
	for _, v := range vals {
		s += v
	}
	return s / float64(len(vals))
}

// Median returns the middle value of a sorted copy.
func Median(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sorted := make([]float64, len(vals))
	copy(sorted, vals)
	// Simple insertion sort for small N
	for i := 1; i < len(sorted); i++ {
		j := i
		for j > 0 && sorted[j-1] > sorted[j] {
			sorted[j-1], sorted[j] = sorted[j], sorted[j-1]
			j--
		}
	}
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}
