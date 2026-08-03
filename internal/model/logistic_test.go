package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFitLogisticSeparable(t *testing.T) {
	t.Parallel()

	// Feature index 0 > 0 -> positive class; easy to separate.
	X := [][]float64{}
	y := []float64{}
	for i := 0; i < 200; i++ {
		X = append(X, []float64{float64(i%4) - 1.5})
		if i%4 >= 2 {
			y = append(y, 1)
		} else {
			y = append(y, 0)
		}
	}

	model, err := FitLogistic(X, y, LogisticOptions{Epochs: 4000, LearnRate: 0.2})
	require.NoError(t, err)

	is := assert.New(t)
	is.True(model.Predict([]float64{0.5}) > 0.7, "positive feature should predict positive class")
	is.True(model.Predict([]float64{-0.5}) < 0.3, "negative feature should predict negative class")
}
