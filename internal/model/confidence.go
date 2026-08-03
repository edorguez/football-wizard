package model

import "math"

// Confidence labels a probability estimate by how far it sits from 0.5.
type Confidence int

const (
	ConfidenceUnknown Confidence = iota
	ConfidenceHigh
	ConfidenceMedium
	ConfidenceLow
)

// FromProbability maps a probability to a confidence band.
//
//	|p - 0.5| >= 0.30 -> High
//	|p - 0.5| >= 0.15 -> Medium
//	otherwise         -> Low
func FromProbability(p float64) Confidence {
	switch d := math.Abs(p - 0.5); {
	case d >= 0.30:
		return ConfidenceHigh
	case d >= 0.15:
		return ConfidenceMedium
	default:
		return ConfidenceLow
	}
}

func (c Confidence) String() string {
	switch c {
	case ConfidenceHigh:
		return "High"
	case ConfidenceMedium:
		return "Medium"
	case ConfidenceLow:
		return "Low"
	default:
		return "Unknown"
	}
}
