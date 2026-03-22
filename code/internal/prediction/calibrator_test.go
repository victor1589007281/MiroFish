package prediction

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalibrator_Calibrate(t *testing.T) {
	c := NewCalibrator()

	// 50% should stay roughly near 50% (with default params A=-1, B=0)
	result := c.Calibrate(0.5)
	assert.InDelta(t, 0.5, result, 0.1)

	// High probability should stay high
	high := c.Calibrate(0.9)
	assert.Greater(t, high, 0.5)
	assert.LessOrEqual(t, high, 0.99)

	// Low probability should stay low
	low := c.Calibrate(0.1)
	assert.Less(t, low, 0.5)
	assert.GreaterOrEqual(t, low, 0.01)
}

func TestCalibrator_Extremization(t *testing.T) {
	c := NewCalibrator()
	c.ExtremizationFactor = 2.0

	// With high extremization, values should be pushed further from 0.5
	result := c.Calibrate(0.7)
	assert.Greater(t, result, 0.5)
}

func TestCalibrator_ClampBounds(t *testing.T) {
	c := NewCalibrator()

	assert.GreaterOrEqual(t, c.Calibrate(0.001), 0.01)
	assert.LessOrEqual(t, c.Calibrate(0.999), 0.99)
}

func TestStdDev(t *testing.T) {
	// All same values -> 0
	assert.InDelta(t, 0.0, StdDev([]float64{0.5, 0.5, 0.5}), 0.001)

	// Single value -> 0
	assert.InDelta(t, 0.0, StdDev([]float64{0.5}), 0.001)

	// Spread values
	sd := StdDev([]float64{0.1, 0.5, 0.9})
	assert.Greater(t, sd, 0.0)
}

func TestMean(t *testing.T) {
	assert.InDelta(t, 0.5, Mean([]float64{0.3, 0.5, 0.7}), 0.001)
	assert.InDelta(t, 0.0, Mean(nil), 0.001)
}

func TestMedian(t *testing.T) {
	assert.InDelta(t, 0.5, Median([]float64{0.1, 0.9, 0.5}), 0.001)
	assert.InDelta(t, 0.5, Median([]float64{0.3, 0.7}), 0.001)
	assert.InDelta(t, 0.0, Median(nil), 0.001)
}

func TestExtremize(t *testing.T) {
	// Factor 1.0 should be identity
	assert.InDelta(t, 0.7, extremize(0.7, 1.0), 0.001)

	// Factor > 1 pushes away from 0.5
	e := extremize(0.7, 2.0)
	assert.Greater(t, e, 0.7)

	// Factor > 1 on low probability pushes lower
	e2 := extremize(0.3, 2.0)
	assert.Less(t, e2, 0.3)
}
