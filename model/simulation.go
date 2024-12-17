package model

import (
	"log"
	"time"

	"github.com/CoralCoralCoralCoral/simulation-engine/logger"
	"github.com/google/uuid"
)

type Simulation struct {
	config            Config
	entity_generator  EntityGenerator
	start_time        time.Time
	epoch             int64
	time_step         int64
	agents            []*Agent
	jurisdictions     []*Jurisdiction
	households        []*Space
	offices           []*Space
	social_spaces     []*Space
	healthcare_spaces []*Space
	is_paused         bool
	should_quit       bool
	commands          chan Command
	logger            logger.Logger
}

func NewSimulation(config Config, entity_generator EntityGenerator) Simulation {
	// note we are creating a logger named logger_ to avoid shadowing the package
	logger_ := logger.NewLogger()

	// attach an internal logger to log processed commands for debugging
	logger_.Subscribe(func(event *logger.Event) {
		switch event.Type {
		case SimulationInitialized:
			log.Print("simulation initialized")
		case CommandProcessed:
			log.Printf("processed command of type %s", event.Payload.(CommandProcessedPayload).Command.Type)
		}
	})

	return Simulation{
		config:           config,
		entity_generator: entity_generator,
		start_time:       time.Now(),
		epoch:            0,
		time_step:        config.TimeStep,
		commands:         make(chan Command),
		logger:           logger_,
	}
}

func (sim *Simulation) Start() {
	sim.initialize()
	sim.infectRandomAgent()

	for {
		if sim.should_quit {
			return
		}

		select {
		case command := <-sim.commands:
			sim.processCommand(command)
		default:
			sim.simulateEpoch()
		}
	}
}

func (sim *Simulation) Subscribe(subscriber func(event *logger.Event)) {
	sim.logger.Subscribe(subscriber)
}

func (sim *Simulation) SendCommand(command Command) {
	sim.commands <- command
}

func (sim *Simulation) Id() uuid.UUID {
	return sim.config.Id
}

func (sim *Simulation) initialize() {
	budgetConfig := InitialiseBudget(&sim.logger)

	// InitialiseBudget(&sim.logger)
	sim.logger.Subscribe(budgetConfig.NewEventSubscriber())

	// start broadcasting logged events to listeners
	go sim.logger.Broadcast()

	sim.generate_entities()

	sim.logger.Log(logger.Event{
		Type: SimulationInitialized,
	})
}

func (sim *Simulation) generate_entities() {
	entities := sim.entity_generator.Generate(&sim.config)

	sim.agents = entities.agents
	sim.jurisdictions = entities.jurisdictions
	sim.households = entities.households
	sim.offices = entities.offices
	sim.social_spaces = entities.social_spaces
	sim.healthcare_spaces = entities.healthcare_spaces
}

func (sim *Simulation) processCommand(command Command) {
	switch command.Type {
	case Quit:
		sim.should_quit = true
	case Pause:
		sim.is_paused = true
	case Resume:
		sim.is_paused = false
	case ApplyPolicyUpdate:
		if payload, ok := command.Payload.(*ApplyPolicyUpdatePayload); ok {
			sim.applyPolicyUpdate(*payload)
		}
	}

	sim.logger.Log(logger.Event{
		Type: CommandProcessed,
		Payload: CommandProcessedPayload{
			Epoch:   sim.epoch,
			Command: command,
		},
	})
}

func (sim *Simulation) simulateEpoch() {
	if sim.is_paused {
		return
	}

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

	for _, healthcare_space := range sim.healthcare_spaces {
		healthcare_space.update(sim)
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

func (sim *Simulation) infectRandomAgent() {
	agent_idx := sampleUniform(0, int64(len(sim.agents)-1))
	sim.agents[agent_idx].infect(sim)
}

func (sim *Simulation) time() time.Time {
	return sim.start_time.Add(time.Duration(sim.epoch*sim.time_step) * time.Millisecond)
}

func (sim *Simulation) applyPolicyUpdate(payload ApplyPolicyUpdatePayload) {
	for _, jur := range sim.jurisdictions {
		if jur.id == payload.JurisdictionId {
			jur.applyPolicyUpdate(sim, &payload)
			return
		}
	}
}
