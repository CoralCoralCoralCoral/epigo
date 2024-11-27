package geo

import (
	"log"
	"time"

	"golang.org/x/exp/rand"
)

type MSOASampler struct {
	msoas        []*MSOA
	total_weight float64
	rng          *rand.Rand
}

func NewMSOASampler() *MSOASampler {
	msoas := loadMSOAs()

	total_weight := 0.0
	for _, msoa := range msoas {
		total_weight += msoa.PopulationDensity
	}

	return &MSOASampler{
		msoas:        msoas,
		total_weight: total_weight,
		rng:          rand.New(rand.NewSource(uint64(time.Now().UnixNano()))),
	}
}

func (sampler *MSOASampler) Sample() *MSOA {
	random_weight := sampler.rng.Float64() * sampler.total_weight

	// Find the corresponding item
	current_weight := 0.0
	for _, msoa := range sampler.msoas {
		current_weight += msoa.PopulationDensity
		if random_weight <= current_weight {
			return msoa
		}
	}

	log.Fatalln("this should be unreachable")
	return nil
}

// SampleMSOA randomly selects a MSOA based on population density.
func SampleMSOA(msoas []MSOA) MSOA {
	var total_weight float64
	weights := make([]float64, len(msoas))

	for i, loc := range msoas {
		total_weight += loc.PopulationDensity
		weights[i] = total_weight
	}

	rng := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
	randomWeight := rng.Float64() * total_weight

	for i, weight := range weights {
		if randomWeight <= weight {
			return msoas[i]
		}
	}

	// Fallback (should not happen)
	return msoas[len(msoas)-1]
}
