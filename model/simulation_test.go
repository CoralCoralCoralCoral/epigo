package model

import (
	"testing"

	"github.com/google/uuid"
)

func TestSimulation(t *testing.T) {
	pathogen := Pathogen{
		IncubationPeriod:           [2]float64{3 * 24 * 60 * 60 * 1000, 8 * 60 * 60 * 1000},
		RecoveryPeriod:             [2]float64{7 * 24 * 60 * 60 * 1000, 8 * 60 * 60 * 1000},
		ImmunityPeriod:             [2]float64{330 * 24 * 60 * 60 * 1000, 90 * 24 * 60 * 60 * 1000},
		PrehospitalizationPeriod:   [2]float64{3 * 24 * 60 * 60 * 1000, 8 * 60 * 60 * 1000},
		HospitalizationPeriod:      [2]float64{7 * 24 * 60 * 60 * 1000, 3 * 24 * 60 * 60 * 1000},
		QuantaEmissionRate:         [2]float64{500, 150},
		HospitalizationProbability: 0.15,
		DeathProbability:           0.75,
		AsymptomaticProbability:    0.10,
	}

	config := Config{
		Id:        uuid.New(),
		TimeStep:  15 * 60 * 1000,
		NumAgents: 1000000,
		Pathogen:  pathogen,
	}

	NewSimulation(config, NewDefaultEntityGenerator())
}
