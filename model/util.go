package model

import (
	"math"

	"gonum.org/v1/gonum/stat/distuv"
)

func sampleBernoulli(p float64) float64 {
	bernoulli := distuv.Binomial{
		N: 1, // N = 1 for a Bernoulli trial
		P: p,
	}

	return bernoulli.Rand()
}

func sampleNormal(mean, sd float64) float64 {
	normalDist := distuv.Normal{
		Mu:    mean, // Mean (µ)
		Sigma: sd,   // Standard deviation (σ)
	}

	// Sample a random value from the normal distribution
	return normalDist.Rand()
}

func sampleUniform(min, max int64) int64 {
	uniDist := distuv.Uniform{
		Min: float64(min),
		Max: float64(max + 1), // We set Max + 1 so the result can include max
	}

	return int64(math.Floor(uniDist.Rand()))
}

// // Calculate the mean of a slice of float64
// func calculateMean(data []float64) float64 {
// 	var sum float64
// 	for _, value := range data {
// 		sum += value
// 	}
// 	return sum / float64(len(data))
// }

// // Calculate the standard deviation of a slice of float64
// func calculateStandardDeviation(data []float64) float64 {
// 	mean := calculateMean(data)
// 	var variance float64
// 	for _, value := range data {
// 		variance += math.Pow(value-mean, 2)
// 	}
// 	variance /= float64(len(data)) // Population SD, use len(data)-1 for sample SD
// 	return math.Sqrt(variance)
// }
