package model

import (
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/umran/epigo/logger"
)

type Simulation struct {
	id              uuid.UUID
	pathogen        Pathogen
	start_time      time.Time
	epoch           int64
	time_step       int64
	agents          []*Agent
	households      []*Space
	offices         []*Space
	social_spaces   []*Space
	is_mask_mandate bool
	is_lockdown     bool
	commands        chan Command
	logger          logger.Logger
}

type Command string

func NewSimulation(agent_count, time_step int64, pathogen_profile Pathogen) Simulation {
	households := createHouseholds(agent_count)
	offices := createOffices(agent_count)
	social_spaces := createSocialSpaces(agent_count / 100)
	agents := createAgents(agent_count, households, offices, social_spaces)

	logger := logger.NewLogger()

	return Simulation{
		id:            uuid.New(),
		pathogen:      pathogen_profile,
		start_time:    time.Now(),
		epoch:         0,
		time_step:     time_step,
		agents:        agents,
		households:    households,
		offices:       offices,
		social_spaces: social_spaces,
		commands:      make(chan Command),
		logger:        logger,
	}
}

func (sim *Simulation) Start() {
	go sim.logger.Broadcast()

	sim.infect_random_agent()

	for {
		select {
		case command := <-sim.commands:
			sim.process_command(command)
		default:
			sim.simulate_epoch()
		}
	}
}

func (sim *Simulation) Subscribe(subscriber func(event *logger.Event)) {
	sim.logger.Subscribe(subscriber)
}

func (sim *Simulation) SendCommand(command Command) {
	sim.commands <- command
}

func (sim *Simulation) process_command(command Command) {
	switch command {
	case "lockdown\n":
		sim.toggle_lockdown()
	case "mask mandate\n":
		sim.toggle_mask_mandate()
	}

	sim.logger.Log(logger.Event{
		Type: CommandProcessed,
		Payload: CommandProcessedPayload{
			Epoch:   sim.epoch,
			Command: command,
		},
	})
}

func (sim *Simulation) simulate_epoch() {
	sim.epoch = sim.epoch + 1

	for _, agent := range sim.agents {
		agent.update(sim)
	}

	for _, household := range sim.households {
		household.update(sim)
	}

	for _, office := range sim.offices {
		office.update(sim)
	}

	for _, social_space := range sim.social_spaces {
		social_space.update(sim)
	}

	sim.logger.Log(logger.Event{
		Type: EpochEnd,
		Payload: EpochEndPayload{
			Epoch:    sim.epoch,
			TimeStep: sim.time_step,
			Time:     sim.time(),
		},
	})
}

func (sim *Simulation) infect_random_agent() {
	agent_idx := sampleUniform(0, int64(len(sim.agents)-1))
	sim.agents[agent_idx].infect(sim)
}

func (sim *Simulation) toggle_mask_mandate() {
	sim.is_mask_mandate = !sim.is_mask_mandate
}

func (sim *Simulation) toggle_lockdown() {
	sim.is_lockdown = !sim.is_lockdown
}

func (sim *Simulation) time() time.Time {
	return sim.start_time.Add(time.Duration(sim.epoch*sim.time_step) * time.Millisecond)
}

func createHouseholds(total_capacity int64) []*Space {
	households := make([]*Space, 0)

	for remaining_capacity := total_capacity; remaining_capacity > 0; {
		capacity := int64(math.Max(math.Floor(sampleNormal(4, 1)), 1))

		if capacity > remaining_capacity {
			capacity = remaining_capacity
		}

		household := newHousehold(capacity)
		households = append(households, &household)

		remaining_capacity -= capacity
	}

	return households
}

func createOffices(total_capacity int64) []*Space {
	offices := make([]*Space, 0)

	for remaining_capacity := total_capacity; remaining_capacity > 0; {
		capacity := int64(math.Max(math.Floor(sampleNormal(10, 2)), 1))

		if capacity > remaining_capacity {
			capacity = remaining_capacity
		}

		office := newOffice(capacity)
		offices = append(offices, &office)

		remaining_capacity -= capacity
	}

	return offices
}

func createSocialSpaces(total_capacity int64) []*Space {
	social_spaces := make([]*Space, 0)

	for remaining_capacity := total_capacity; remaining_capacity > 0; {
		capacity := int64(math.Max(math.Floor(sampleNormal(10, 2)), 1))

		if capacity > remaining_capacity {
			capacity = remaining_capacity
		}

		social_space := newSocialSpace(capacity)
		social_spaces = append(social_spaces, &social_space)

		remaining_capacity -= capacity
	}

	return social_spaces
}

func createAgents(count int64, households, offices []*Space, social_spaces []*Space) []*Agent {
	agents := make([]*Agent, count)

	for i := 0; i < int(count); i++ {
		agent := newAgent()

		num_social_spaces := int(math.Max(1, math.Floor(sampleNormal(5, 4))))
		for i := 0; i < num_social_spaces; i++ {
			agent.social_spaces = append(agent.social_spaces, social_spaces[sampleUniform(0, int64(len(social_spaces)-1))])
		}

		agents[i] = &agent
	}

	// allocate agents to households
	household_idx, household_allocated_capacity := 0, 0
	for _, agent := range agents {
		household := households[household_idx]
		agent.household = household
		agent.location = household

		household_allocated_capacity += 1
		if household_allocated_capacity == int(household.capacity) {
			household_idx += 1
			household_allocated_capacity = 0
		}
	}

	// allocate agents to offices
	office_idx, office_allocated_capacity := 0, 0
	for _, agent := range agents {
		office := offices[office_idx]
		agent.office = office

		office_allocated_capacity += 1
		if office_allocated_capacity == int(office.capacity) {
			office_idx += 1
			office_allocated_capacity = 0
		}
	}

	return agents
}
