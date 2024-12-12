package model

import (
	"log"
	"math"
	"math/rand"

	"github.com/CoralCoralCoralCoral/simulation-engine/geo"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/xy"
)

type DefaultEntityGenerator struct {
}

func NewDefaultEntityGenerator() *DefaultEntityGenerator {
	return &DefaultEntityGenerator{}
}

func (g *DefaultEntityGenerator) Generate(config *Config) Entities {
	msoa_sampler := geo.NewMSOASampler()

	jurisdictions := jurisdictionsFromFeatures()
	households := createHouseholds(config.NumAgents, jurisdictions, msoa_sampler)
	offices := createOffices(config.NumAgents, jurisdictions, msoa_sampler)
	social_spaces := createSocialSpaces(config.NumAgents/100, jurisdictions, msoa_sampler)
	healthcare_spaces := createHealthCareSpaces(config.NumAgents/1000*5, jurisdictions, msoa_sampler)
	agents := createAgents(config.NumAgents, households, offices, social_spaces, healthcare_spaces)

	return Entities{
		agents,
		jurisdictions,
		households,
		offices,
		social_spaces,
		healthcare_spaces,
	}
}

func createHouseholds(total_capacity int64, jurisdictions []*Jurisdiction, msoa_sampler *geo.MSOASampler) []*Space {
	households := make([]*Space, 0)

	for remaining_capacity := total_capacity; remaining_capacity > 0; {
		capacity := int64(math.Max(math.Floor(sampleNormal(4, 1)), 1))

		if capacity > remaining_capacity {
			capacity = remaining_capacity
		}

		household := newHousehold(capacity)
		household.jurisdiction = sampleJurisdiction(jurisdictions, msoa_sampler)
		households = append(households, &household)

		remaining_capacity -= capacity
	}

	return households
}

func createOffices(total_capacity int64, jurisdictions []*Jurisdiction, msoa_sampler *geo.MSOASampler) []*Space {
	offices := make([]*Space, 0)

	for remaining_capacity := total_capacity; remaining_capacity > 0; {
		capacity := int64(math.Max(math.Floor(sampleNormal(10, 2)), 1))

		if capacity > remaining_capacity {
			capacity = remaining_capacity
		}

		office := newOffice(capacity)
		office.jurisdiction = sampleJurisdiction(jurisdictions, msoa_sampler)
		offices = append(offices, &office)

		remaining_capacity -= capacity
	}

	return offices
}

func createSocialSpaces(total_capacity int64, jurisdictions []*Jurisdiction, msoa_sampler *geo.MSOASampler) []*Space {
	social_spaces := make([]*Space, 0)

	for remaining_capacity := total_capacity; remaining_capacity > 0; {
		capacity := int64(math.Max(math.Floor(sampleNormal(10, 2)), 1))

		if capacity > remaining_capacity {
			capacity = remaining_capacity
		}

		social_space := newSocialSpace(capacity)
		social_space.jurisdiction = sampleJurisdiction(jurisdictions, msoa_sampler)
		social_spaces = append(social_spaces, &social_space)

		remaining_capacity -= capacity
	}

	return social_spaces
}

func createHealthCareSpaces(total_capacity int64, jurisdictions []*Jurisdiction, msoa_sampler *geo.MSOASampler) []*Space {
	healthcare_spaces := make([]*Space, 0)

	for remaining_capacity := total_capacity; remaining_capacity > 0; {
		capacity := int64(math.Max(math.Floor(sampleNormal(173, 25)), 1))

		if capacity > remaining_capacity {
			capacity = remaining_capacity
		}

		healthcare_space := newHealthcareSpace(capacity)
		healthcare_space.jurisdiction = sampleJurisdiction(jurisdictions, msoa_sampler)
		healthcare_spaces = append(healthcare_spaces, &healthcare_space)

		remaining_capacity -= capacity
	}

	return healthcare_spaces
}

func createAgents(count int64, households, offices []*Space, social_spaces []*Space, healthcare_spaces []*Space) []*Agent {
	agents := make([]*Agent, count)

	office_sampler := distanceWeighted(offices)
	social_space_sampler := distanceWeighted(social_spaces)
	healthcare_space_sampler := distanceWeighted(healthcare_spaces)

	household_idx, household_allocated_capacity := 0, 0
	for i := 0; i < int(count); i++ {
		household := households[household_idx]
		agents[i] = createAgent(household, office_sampler, social_space_sampler, healthcare_space_sampler)

		household_allocated_capacity += 1
		if household_allocated_capacity == int(household.capacity) {
			household_idx += 1
			household_allocated_capacity = 0
		}
	}

	return agents
}

func createAgent(household *Space, office_sampler, social_space_sampler, healthcare_space_sampler func(space *Space) *Space) *Agent {
	agent := newAgent()
	agent.household = household
	agent.location = household

	// create distance matrix from agent to offices
	agent.office = office_sampler(agent.household)

	// create distance matrix from agent to social spaces
	num_social_spaces := int(math.Max(1, math.Floor(sampleNormal(5, 4))))
	for i := 0; i < num_social_spaces; i++ {
		agent.social_spaces = append(agent.social_spaces, social_space_sampler(agent.household))
	}

	// create distance matrix from agent to offices
	num_healthcare_spaces := int(math.Max(1, math.Floor(sampleNormal(5, 4))))
	for i := 0; i < num_healthcare_spaces; i++ {
		agent.healthcare_spaces = append(agent.healthcare_spaces, healthcare_space_sampler(agent.household))
	}

	return &agent
}

func distanceWeighted(spaces []*Space) func(space *Space) *Space {
	weights_map := make(map[string][]float64)

	return func(space *Space) *Space {
		if weights, ok := weights_map[space.jurisdiction.id]; ok {
			return randomWeightedSample(spaces, weights)
		}

		weights := calculateWeights(spaces, space)
		weights_map[space.jurisdiction.id] = weights

		return randomWeightedSample(spaces, weights)
	}
}

// randomWeightedSample selects an space based on weights
func randomWeightedSample(spaces []*Space, weights []float64) *Space {
	total_weight := 0.0
	for _, weight := range weights {
		total_weight += weight
	}

	random_weight := rand.Float64() * total_weight

	// Find the corresponding item
	current_weight := 0.0
	for i, weight := range weights {
		current_weight += weight
		if random_weight <= current_weight {
			return spaces[i]
		}
	}

	log.Fatalf("this should be unreachable: current weight: %f total_weight %f", current_weight, total_weight)
	return nil
}

// calculateWeights calculates weights for areas based on distances to the agent's home area
func calculateWeights(spaces []*Space, origin *Space) []float64 {
	weights := make([]float64, len(spaces))
	origin_point, err := xy.Centroid(origin.jurisdiction.feature.Geometry)

	if err != nil {
		log.Fatalf("failed to calculate centroid of jurisdiction feature geometry")
	}

	for i, space := range spaces {
		space_point, err := xy.Centroid(space.jurisdiction.feature.Geometry)
		if err != nil {
			log.Fatalf("failed to calculate centroid of jurisdiction feature geometry")
		}

		distance := calculateDistance(origin_point, space_point)
		weights[i] = 1 / (distance + 1e-6) // Avoid division by zero
	}

	return weights
}

func calculateDistance(p1, p2 geom.Coord) float64 {
	if len(p1) < 2 || len(p2) < 2 {
		log.Fatalf("Invalid coordinates: %v, %v", p1, p2)
	}

	dx := p1[0] - p2[0]
	dy := p1[1] - p2[1]

	return math.Sqrt(dx*dx + dy*dy)
}

func jurisdictionsFromFeatures() []*Jurisdiction {
	features := geo.LoadFeatures()

	// allocate array of length feature length + 1 (to also contain the GLOBAL jurisdiction)
	jurisdictions := make([]*Jurisdiction, 0, len(features)+1)

	// create jurisdictions
	for _, feature := range features {
		jurisdictions = append(jurisdictions, newJurisdiction(feature.Code(), nil, feature))
	}

	// assign parents
	for _, feature := range features {
		if parent_code := feature.ParentCode(); parent_code != "" {
			jur := findJurisdiction(jurisdictions, func(val *Jurisdiction) bool {
				return val.Id() == feature.Code()
			})

			// find parent_jurisdiction
			parent_jur := findJurisdiction(jurisdictions, func(val *Jurisdiction) bool {
				return val.Id() == parent_code
			})

			if jur != nil && parent_jur != nil {
				jur.parent = parent_jur
			}
		}
	}

	// assign the highest level jurisdictions (orphan jurisdictions to the GLOBAL jurisdiction)
	global_jur := newJurisdiction("GLOBAL", nil, nil)
	for _, jur := range jurisdictions {
		if jur.parent == nil {
			jur.parent = global_jur
		}
	}

	jurisdictions = append(jurisdictions, global_jur)

	return jurisdictions
}

func sampleJurisdiction(jurisdictions []*Jurisdiction, msoa_sampler *geo.MSOASampler) *Jurisdiction {
	msoa := msoa_sampler.Sample()

	jur := findJurisdiction(jurisdictions, func(val *Jurisdiction) bool {
		return val.Id() == msoa.GISCode
	})

	return jur
}

func findJurisdiction(jurisdictions []*Jurisdiction, predicate func(value *Jurisdiction) bool) *Jurisdiction {
	for _, value := range jurisdictions {
		if predicate(value) {
			return value
		}
	}

	return nil
}
