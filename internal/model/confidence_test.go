package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromProbability(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    float64
		want Confidence
	}{
		{name: "near certain", p: 0.95, want: ConfidenceHigh},
		{name: "strong", p: 0.80, want: ConfidenceHigh},
		{name: "medium", p: 0.70, want: ConfidenceMedium},
		{name: "weak", p: 0.60, want: ConfidenceLow},
		{name: "coin flip", p: 0.50, want: ConfidenceLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FromProbability(tt.p))
		})
	}
}
