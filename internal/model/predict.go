package model

// MarketPrediction is a single probability estimate for one market.
type MarketPrediction struct {
	Market      Market
	Outcome     string
	Line        float64
	Probability float64
	Confidence  Confidence
}

// MatchPrediction bundles the main 1X2 / goals forecast with every market.
type MatchPrediction struct {
	HomeTeam           string
	AwayTeam           string
	HomeWin            float64
	Draw               float64
	AwayWin            float64
	ExpectedHomeGoals  float64
	ExpectedAwayGoals  float64
	ExpectedTotalGoals float64
	Markets            []MarketPrediction
}

// FirstScorerModel holds the two classifiers that decide who scores first.
type FirstScorerModel struct {
	Home *LogisticModel
	Away *LogisticModel
}

// Predictor is a fully trained engine ready to forecast fixtures.
type Predictor struct {
	poisson     *PoissonModel
	logistic    map[Market]*LogisticModel
	firstScorer *FirstScorerModel
	thresholds  map[Market]float64
	goalLines   []float64
	engine      *Engine
}

// PredictMatch forecasts a fixture using the end-of-training form state.
func (p *Predictor) PredictMatch(homeID, awayID uint, homeName, awayName string) *MatchPrediction {
	homeG, awayG := p.poisson.ExpectedGoals(homeID, awayID)
	features := p.engine.Features(homeID, awayID)

	pred := &MatchPrediction{
		HomeTeam:           homeName,
		AwayTeam:           awayName,
		HomeWin:            p.poisson.ProbHomeWin(homeID, awayID),
		Draw:               p.poisson.ProbDraw(homeID, awayID),
		AwayWin:            p.poisson.ProbAwayWin(homeID, awayID),
		ExpectedHomeGoals:  homeG,
		ExpectedAwayGoals:  awayG,
		ExpectedTotalGoals: homeG + awayG,
	}

	pred.Markets = append(pred.Markets, p.goalLineMarkets(homeID, awayID)...)
	pred.Markets = append(pred.Markets, p.binaryMarkets(features)...)
	pred.Markets = append(pred.Markets, p.firstScorerMarkets(features)...)

	return pred
}

func (p *Predictor) goalLineMarkets(homeID, awayID uint) []MarketPrediction {
	lines := p.goalLines
	if len(lines) == 0 {
		lines = DefaultGoalLines
	}

	markets := make([]MarketPrediction, 0, len(lines))
	for _, line := range lines {
		prob := p.poisson.ProbOver(homeID, awayID, line)
		markets = append(markets, MarketPrediction{
			Market:      MarketTotalGoals,
			Outcome:     "Over",
			Line:        line,
			Probability: prob,
			Confidence:  FromProbability(prob),
		})
	}
	return markets
}

func (p *Predictor) binaryMarkets(features *Features) []MarketPrediction {
	vector := features.Vector()

	var markets []MarketPrediction

	for _, spec := range binaryMarketSpecs {
		model := p.logistic[spec.Market]
		if model == nil {
			continue
		}
		prob := model.Predict(vector)
		markets = append(markets, MarketPrediction{
			Market:      spec.Market,
			Outcome:     marketOutcome(spec.Market),
			Line:        p.line(spec.Market),
			Probability: prob,
			Confidence:  FromProbability(prob),
		})
	}

	return markets
}

func marketOutcome(m Market) string {
	switch m {
	case MarketBTTS:
		return "Yes"
	case MarketHomeFirstHalf:
		return "Home scores in first half"
	case MarketAwayFirstHalf:
		return "Away scores in first half"
	case MarketHomeSecondHalf:
		return "Home scores in second half"
	case MarketAwaySecondHalf:
		return "Away scores in second half"
	default:
		return "Over"
	}
}

func (p *Predictor) firstScorerMarkets(features *Features) []MarketPrediction {
	vector := features.Vector()

	if p.firstScorer == nil || (p.firstScorer.Home == nil && p.firstScorer.Away == nil) {
		return nil
	}

	homeProb := 0.0
	if p.firstScorer.Home != nil {
		homeProb = p.firstScorer.Home.Predict(vector)
	}
	awayProb := 0.0
	if p.firstScorer.Away != nil {
		awayProb = p.firstScorer.Away.Predict(vector)
	}

	neither := 1 - homeProb - awayProb
	if neither < 0 {
		sum := homeProb + awayProb
		if sum > 0 {
			homeProb /= sum
			awayProb /= sum
		}
		neither = 0
	}

	outcomes := []struct {
		outcome string
		prob    float64
	}{
		{"Home first", homeProb},
		{"Away first", awayProb},
		{"Neither", neither},
	}

	markets := make([]MarketPrediction, 0, len(outcomes))
	for _, o := range outcomes {
		markets = append(markets, MarketPrediction{
			Market:      MarketFirstScorer,
			Outcome:     o.outcome,
			Probability: o.prob,
			Confidence:  FromProbability(o.prob),
		})
	}
	return markets
}

func (p *Predictor) line(market Market) float64 {
	if v, ok := p.thresholds[market]; ok {
		return v
	}
	return DefaultThresholds[market]
}
