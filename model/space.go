package model

import (
	"math"

	"github.com/CoralCoralCoralCoral/simulation-engine/logger"
	"github.com/google/uuid"
)

const TestSensitivity float64 = 0.7
const TestSpecificity float64 = 0.999

const Household SpaceType = "household"
const Office SpaceType = "office"
const SocialSpace SpaceType = "social_space"
const HealthCareSpace SpaceType = "healthcare_space"

type Space struct {
	id                     uuid.UUID
	type_                  SpaceType
	jurisdiction           *Jurisdiction
	occupants              []*Agent
	capacity               int64
	volume                 float64
	air_change_rate        float64
	total_infectious_doses float64

	// healthcare related props
	test_capacity int64
	test_backlog  chan bool
}

type SpaceType string

func newHousehold(capacity int64) Space {
	return Space{
		id:                     uuid.New(),
		type_:                  Household,
		jurisdiction:           nil,
		occupants:              make([]*Agent, 0),
		capacity:               capacity,
		volume:                 sampleNormal(17, 2),
		air_change_rate:        sampleNormal(7, 1),
		total_infectious_doses: 0,
	}
}

func newOffice(capacity int64) Space {
	return Space{
		id:                     uuid.New(),
		type_:                  Office,
		jurisdiction:           nil,
		occupants:              make([]*Agent, 0),
		capacity:               capacity,
		volume:                 sampleNormal(60, 20),
		air_change_rate:        sampleNormal(20, 5),
		total_infectious_doses: 0,
	}
}

func newSocialSpace(capacity int64) Space {
	return Space{
		id:                     uuid.New(),
		type_:                  SocialSpace,
		jurisdiction:           nil,
		occupants:              make([]*Agent, 0),
		capacity:               capacity,
		volume:                 sampleNormal(60, 10),
		air_change_rate:        sampleNormal(20, 5),
		total_infectious_doses: 0,
	}
}

func newHealthcareSpace(capacity int64) Space {
	return Space{
		id:                     uuid.New(),
		type_:                  HealthCareSpace,
		jurisdiction:           nil,
		occupants:              make([]*Agent, 0),
		capacity:               capacity,
		volume:                 sampleNormal(60, 10),
		air_change_rate:        sampleNormal(20, 5),
		total_infectious_doses: 0,

		test_capacity: int64(math.Max(1, math.Floor(sampleNormal(300, 150)))),
		test_backlog:  make(chan bool, 1000000),
	}
}

func (space *Space) update(sim *Simulation) {
	policy := space.resolvePolicy()

	// introduce new infectious doses from infectious occupants
	for _, occupant := range space.occupants {
		if occupant.state == Infectious {
			filtration_efficiency := 0.0
			if policy != nil && policy.IsMaskMandate && occupant.is_compliant {
				filtration_efficiency = occupant.mask_filtration_efficiency
			}

			quanta_emission_rate := (1 - filtration_efficiency) * occupant.infection_profile.quanta_emission_rate / 3600
			space.total_infectious_doses += quanta_emission_rate * float64(sim.time_step) / 1000
		}
	}

	// remove infectious doses due to ventilation
	space.total_infectious_doses = space.total_infectious_doses * math.Exp(-1*(space.air_change_rate/3600)*float64(sim.time_step)/1000)

	// if it is the end of a day and the space is a healthcare space, report test results
	if space.type_ == HealthCareSpace && (sim.epoch*sim.time_step)%(24*60*60*1000) == 0 {
		space.dispatchTestingUpdateEvent(sim)
	}
}

func (space *Space) addAgent(sim *Simulation, agent *Agent) {
	space.occupants = append(space.occupants, agent)

	policy := space.resolvePolicy()
	if space.type_ == HealthCareSpace && policy != nil {
		switch policy.TestStrategy {
		case TestEveryone:
			if agent.infection_profile != nil {
				if sampleBernoulli(TestSensitivity) == 1 {
					space.test_backlog <- true
				} else {
					space.test_backlog <- false
				}
			} else { // agent is not infected, so simulate test using test specificity
				if sampleBernoulli(TestSpecificity) == 1 {
					space.test_backlog <- false
				} else {
					space.test_backlog <- true
				}
			}
		case TestSymptomatic:
			if agent.infection_profile != nil && !agent.infection_profile.is_asymptomatic {
				if sampleBernoulli(TestSensitivity) == 1 {
					space.test_backlog <- true
				} else {
					space.test_backlog <- false
				}
			}
		}
	}

	space.dispatchOccupancyUpdateEvent(sim)
}

func (space *Space) removeAgent(sim *Simulation, agent *Agent) {
	for idx, candidate := range space.occupants {
		if candidate.id == agent.id {
			space.occupants = append(space.occupants[:idx], space.occupants[idx+1:]...)
			break
		}
	}

	space.dispatchOccupancyUpdateEvent(sim)
}

func (space *Space) dispatchTestingUpdateEvent(sim *Simulation) {
	test_capacity := space.test_capacity
	if policy := space.resolvePolicy(); policy != nil {
		if policy.TestCapacityMultiplier > 0 {
			test_capacity = int64(math.Ceil(float64(test_capacity) * policy.TestCapacityMultiplier))
		}
	}

	positives := 0
	negatives := 0

loop:
	for i := 0; i < int(test_capacity); i++ {
		select {
		case result := <-space.test_backlog:
			if result {
				positives += 1
			} else {
				negatives += 1
			}
		default:
			break loop
		}
	}

	backlog := len(space.test_backlog)

	event := logger.Event{
		Type: SpaceTestingUpdate,
		Payload: SpaceTestingUpdatePayload{
			Epoch:     sim.epoch,
			Positives: int64(positives),
			Negatives: int64(negatives),
			Backlog:   int64(backlog),
			Capacity:  test_capacity,

			jurisdiction: space.jurisdiction,
		},
	}

	sim.logger.Log(event)
}

func (space *Space) dispatchOccupancyUpdateEvent(sim *Simulation) {
	occupants := make([]struct {
		Id    uuid.UUID  `json:"id"`
		State AgentState `json:"state"`
	}, len(space.occupants))

	for _, occupant := range space.occupants {
		occupants = append(occupants, struct {
			Id    uuid.UUID  `json:"id"`
			State AgentState `json:"state"`
		}{
			Id:    occupant.id,
			State: occupant.state,
		})
	}

	event := logger.Event{
		Type: SpaceOccupancyUpdate,
		Payload: SpaceOccupancyUpdatePayload{
			Epoch:     sim.epoch,
			Id:        space.id,
			Occupants: occupants,
		},
	}

	sim.logger.Log(event)
}

func (space *Space) state() (SpaceType, float64, float64, float64, *Policy) {
	return space.type_, space.volume, space.air_change_rate, space.total_infectious_doses, space.resolvePolicy()
}

func (space *Space) resolvePolicy() (policy *Policy) {
	if space.jurisdiction != nil {
		policy = space.jurisdiction.resolvePolicy()
	}

	return
}
