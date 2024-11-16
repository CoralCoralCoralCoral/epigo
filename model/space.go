package model

import (
	"math"
	"sync"

	"github.com/google/uuid"
	"github.com/umran/epigo/logger"
)

const Household SpaceType = "household"
const Office SpaceType = "office"
const SocialSpace SpaceType = "social_space"

type Space struct {
	mu                     *sync.RWMutex
	id                     uuid.UUID
	type_                  SpaceType
	occupants              []*Agent
	capacity               int64
	volume                 float64
	air_change_rate        float64
	total_infectious_doses float64
}

type SpaceType string

func newHousehold(capacity int64) Space {
	return Space{
		mu:                     new(sync.RWMutex),
		id:                     uuid.New(),
		type_:                  Household,
		occupants:              make([]*Agent, 0),
		capacity:               capacity,
		volume:                 sampleNormal(17, 2),
		air_change_rate:        sampleNormal(7, 1),
		total_infectious_doses: 0,
	}
}

func newOffice(capacity int64) Space {
	return Space{
		mu:                     new(sync.RWMutex),
		id:                     uuid.New(),
		type_:                  Office,
		occupants:              make([]*Agent, 0),
		capacity:               capacity,
		volume:                 sampleNormal(60, 20),
		air_change_rate:        sampleNormal(20, 5),
		total_infectious_doses: 0,
	}
}

func newSocialSpace(capacity int64) Space {
	return Space{
		mu:                     new(sync.RWMutex),
		id:                     uuid.New(),
		type_:                  SocialSpace,
		occupants:              make([]*Agent, 0),
		capacity:               capacity,
		volume:                 sampleNormal(60, 10),
		air_change_rate:        sampleNormal(20, 5),
		total_infectious_doses: 0,
	}
}

func (space *Space) update(sim *Simulation) {
	space.mu.Lock()
	defer space.mu.Unlock()

	// introduce new infectious doses from infectious occupants
	for _, occupant := range space.occupants {
		if occupant.state == Infectious {
			filtration_efficiency := 0.0
			if sim.is_mask_mandate && occupant.is_compliant {
				filtration_efficiency = occupant.mask_filtration_efficiency
			}

			quanta_emission_rate := (1 - filtration_efficiency) * occupant.infection_profile.quanta_emission_rate / 3600
			space.total_infectious_doses += quanta_emission_rate * float64(sim.time_step) / 1000
		}
	}

	// remove infectious doses due to ventilation
	space.total_infectious_doses = space.total_infectious_doses * math.Exp(-1*(space.air_change_rate/3600)*float64(sim.time_step)/1000)
}

func (space *Space) addAgent(sim *Simulation, agent *Agent) {
	space.mu.Lock()
	defer space.mu.Unlock()

	space.occupants = append(space.occupants, agent)

	space.dispatchOccupancyUpdateEvent(sim)
}

func (space *Space) removeAgent(sim *Simulation, agent *Agent) {
	space.mu.Lock()
	defer space.mu.Unlock()

	for idx, candidate := range space.occupants {
		if candidate.id == agent.id {
			space.occupants = append(space.occupants[:idx], space.occupants[idx+1:]...)
			break
		}
	}

	space.dispatchOccupancyUpdateEvent(sim)
}

func (space *Space) dispatchOccupancyUpdateEvent(sim *Simulation) {
	occupants := make([]struct {
		Id    uuid.UUID
		State AgentState
	}, len(space.occupants))

	for _, occupant := range space.occupants {
		occupants = append(occupants, struct {
			Id    uuid.UUID
			State AgentState
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

func (space *Space) state() (float64, float64, float64) {
	space.mu.RLock()
	defer space.mu.RUnlock()

	return space.volume, space.air_change_rate, space.total_infectious_doses
}
