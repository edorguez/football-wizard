package model

import (
	"math"

	"github.com/edorguez/football-wizard/pkg/utils"
)

type LogisticModel struct {
	weights []float64
	trained bool
}

func NewLogisticModel() *LogisticModel {
	return &LogisticModel{}
}

func sigmoid(z float64) float64 {
	return 1.0 / (1.0 + math.Exp(-z))
}

func (m *LogisticModel) Train(features [][]float64, labels []float64, iterations int, learningRate float64) {
	if len(features) == 0 {
		return
	}

	nSamples := len(features)
	nFeatures := len(features[0])
	m.weights = make([]float64, nFeatures)

	for iter := 0; iter < iterations; iter++ {
		gradient := make([]float64, nFeatures)

		for i := 0; i < nSamples; i++ {
			pred := m.PredictRaw(features[i])
			err := pred - labels[i]
			for j := 0; j < nFeatures; j++ {
				gradient[j] += err * features[i][j]
			}
		}

		for j := 0; j < nFeatures; j++ {
			m.weights[j] -= learningRate * gradient[j] / float64(nSamples)
		}
	}

	m.trained = true
}

func (m *LogisticModel) PredictRaw(features []float64) float64 {
	z := 0.0
	for i, w := range m.weights {
		z += w * features[i]
	}
	return sigmoid(z)
}

func (m *LogisticModel) Predict(features []float64) float64 {
	return utils.Round(m.PredictRaw(features), 4)
}

func (m *LogisticModel) IsTrained() bool {
	return m.trained
}
