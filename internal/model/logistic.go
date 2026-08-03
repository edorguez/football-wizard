package model

import (
	"fmt"
	"math"
)

// LogisticOptions configures the gradient-descent fitting procedure.
type LogisticOptions struct {
	Epochs    int
	LearnRate float64
	L2        float64
}

func (o *LogisticOptions) withDefaults() {
	if o.Epochs <= 0 {
		o.Epochs = 2000
	}
	if o.LearnRate <= 0 {
		o.LearnRate = 0.05
	}
	if o.L2 < 0 {
		o.L2 = 0
	}
}

// LogisticModel is a binary classifier trained by gradient descent on
// standardized features. Weights includes the bias as its last element.
type LogisticModel struct {
	Weights []float64
	Mean    []float64
	Std     []float64
}

// FitLogistic fits a regularized logistic regression on the given features
// and binary labels (0/1). Samples with zero variance in a feature collapse
// to a 1.0 standard deviation to keep the fit stable.
func FitLogistic(features [][]float64, labels []float64, opts LogisticOptions) (*LogisticModel, error) {
	opts.withDefaults()

	if len(features) == 0 {
		return nil, fmt.Errorf("no training samples")
	}
	n, d := len(features), len(features[0])
	if len(labels) != n {
		return nil, fmt.Errorf("labels and features length mismatch: %d vs %d", len(labels), n)
	}

	mean := make([]float64, d)
	for _, f := range features {
		for j := range mean {
			mean[j] += f[j]
		}
	}
	for j := range mean {
		mean[j] /= float64(n)
	}

	std := make([]float64, d)
	for _, f := range features {
		for j := range std {
			diff := f[j] - mean[j]
			std[j] += diff * diff
		}
	}
	for j := range std {
		std[j] = math.Sqrt(std[j] / float64(n))
		if std[j] == 0 {
			std[j] = 1
		}
	}

	// Standardize with a bias column appended.
	X := make([][]float64, n)
	for i, f := range features {
		row := make([]float64, d+1)
		for j := range f {
			row[j] = (f[j] - mean[j]) / std[j]
		}
		row[d] = 1
		X[i] = row
	}

	w := make([]float64, d+1)
	for epoch := 0; epoch < opts.Epochs; epoch++ {
		grad := make([]float64, d+1)
		for i, row := range X {
			error := sigmoid(dot(w, row)) - labels[i]
			for j := range grad {
				grad[j] += error * row[j]
			}
		}

		for j := range grad {
			grad[j] /= float64(n)
			if j < d {
				grad[j] += opts.L2 * w[j]
			}
			w[j] -= opts.LearnRate * grad[j]
		}
	}

	return &LogisticModel{Weights: w, Mean: mean, Std: std}, nil
}

// Predict returns the probability of the positive class (1).
func (m *LogisticModel) Predict(features []float64) float64 {
	if m == nil {
		return 0
	}
	sum := m.Weights[len(m.Weights)-1]
	for j := range m.Mean {
		sum += m.Weights[j] * (features[j] - m.Mean[j]) / m.Std[j]
	}
	return sigmoid(sum)
}

func sigmoid(x float64) float64 {
	if x >= 0 {
		z := math.Exp(-x)
		return 1 / (1 + z)
	}
	z := math.Exp(x)
	return z / (1 + z)
}

func dot(a, b []float64) float64 {
	var sum float64
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}
