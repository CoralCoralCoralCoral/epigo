package model

import (
	"math"

	"github.com/CoralCoralCoralCoral/simulation-engine/logger"
	"github.com/google/uuid"
)

const Susceptible AgentState = "susceptible"
const Infected AgentState = "infected"
const Infectious AgentState = "infectious"
const Hospitalized AgentState = "hospitalized"
const Dead AgentState = "dead"
const Immune AgentState = "immune"

type Agent struct {
	id                         uuid.UUID
	household                  *Space
	office                     *Space
	social_spaces              []*Space
	healthcare_spaces          []*Space
	location                   *Space
	location_change_epoch      int64
	next_move_epoch            int64
	state                      AgentState
	state_change_epoch         int64
	infection_profile          *InfectionProfile
	pulmonary_ventilation_rate float64
	is_compliant               bool
	seeks_treatment            bool
	mask_filtration_efficiency float64
}

type AgentState string

func newAgent(config *Config) Agent {
	is_compliant := false
	if sampleBernoulli(config.ComplianceProbability) == 1 {
		is_compliant = true
	}

	seeks_treatment := false
	if sampleBernoulli(config.SeeksTreatmentProbability) == 1 {
		seeks_treatment = true
	}

	return Agent{
		id:                         uuid.New(),
		household:                  nil,
		office:                     nil,
		social_spaces:              make([]*Space, 0),
		healthcare_spaces:          make([]*Space, 0),
		location:                   nil,
		location_change_epoch:      0,
		next_move_epoch:            0,
		state:                      Susceptible,
		state_change_epoch:         0,
		infection_profile:          nil,
		pulmonary_ventilation_rate: sampleNormal(config.PulmonaryVentilationRateMean, config.PulmonaryVentilationRateSd),
		is_compliant:               is_compliant,
		mask_filtration_efficiency: math.Max(sampleNormal(config.MaskFiltrationEfficiencyMean, config.MaskFiltrationEfficiencySd), 0.95),
		seeks_treatment:            seeks_treatment,
	}
}

func (agent *Agent) update(sim *Simulation) {
	if agent.state == Dead {
		return
	}

	agent.updateState(sim)
	agent.updateLocation(sim)
}

func (agent *Agent) updateState(sim *Simulation) {
	state_duration := float64((sim.epoch - agent.state_change_epoch) * sim.time_step)

	switch agent.state {
	case Susceptible:
		is_infected := sampleBernoulli(agent.pInfected(sim))

		if is_infected == 1 {
			agent.infect(sim)
		}
	case Infected:
		if state_duration >= agent.infection_profile.incubation_period {
			agent.setState(sim, Infectious)
		}
	case Infectious:
		switch agent.infection_profile.is_hospitalized {
		case true:
			if state_duration >= agent.infection_profile.prehospitalization_period {
				agent.setState(sim, Hospitalized)
			}
		case false:
			if state_duration >= agent.infection_profile.recovery_period {
				agent.setState(sim, Immune)
			}
		}
	case Hospitalized:
		if state_duration >= agent.infection_profile.hospitalization_period {
			if agent.infection_profile.is_dead {
				agent.setState(sim, Dead)
			} else {
				agent.setState(sim, Immune)
			}
		}
	case Immune:
		if state_duration >= agent.infection_profile.immunity_period {
			agent.infection_profile = nil
			agent.setState(sim, Susceptible)
		}
	case Dead:
		// noop
	default:
		panic("this shouldn't be possible")
	}
}

func (agent *Agent) setState(sim *Simulation, state AgentState) {
	previous_state := agent.state

	agent.state = state
	agent.state_change_epoch = sim.epoch
	agent.dispatchStateUpdateEvent(sim, previous_state)
}

func (agent *Agent) updateLocation(sim *Simulation) {
	if agent.next_move_epoch == 0 {
		// assumes agent is in household
		agent.next_move_epoch = sim.epoch + int64(math.Ceil(sampleNormal(12*60*60*1000, 4*60*60*1000)/float64(sim.time_step)))
	}

	// in the special case where the agent state transitioned to
	// Hospitalized in this epoch, the agent moves to a healthcare space
	// for a duration of hospitalization_period
	if agent.state == Hospitalized && agent.state_change_epoch == sim.epoch {
		agent.setLocation(
			sim,
			agent.healthcare_spaces[sampleUniform(0, int64(len(agent.healthcare_spaces)-1))],
			agent.infection_profile.hospitalization_period,
		)

		return
	}

	// in the special case that the agent state transitioned to Infectious
	// in this epoch and the agent is symptomatic and seeks treatment
	// the agent moves to a healthcare space for a short duration
	if agent.state == Infectious && !agent.infection_profile.is_asymptomatic && agent.seeks_treatment && agent.state_change_epoch == sim.epoch {
		agent.setLocation(
			sim,
			agent.healthcare_spaces[sampleUniform(0, int64(len(agent.healthcare_spaces)-1))],
			sampleNormal(45*60*1000, 15*60*1000),
		)

		return
	}

	if sim.epoch < agent.next_move_epoch {
		return
	}

	location_type, _, _, _, policy := agent.location.state()
	switch location_type {
	case Household:
		if policy != nil && policy.is_lockdown && agent.is_compliant {
			break
		}

		if sampleBernoulli(0.55) == 1 {
			agent.setLocation(
				sim,
				agent.office,
				sampleNormal(8*60*60*1000, 2*60*60*1000),
			)
		} else if sampleBernoulli(0.001) == 1 {
			// simulate randomly going to a healthcare space
			agent.setLocation(
				sim,
				agent.healthcare_spaces[sampleUniform(0, int64(len(agent.healthcare_spaces)-1))],
				sampleNormal(45*60*1000, 15*60*1000),
			)
		} else {
			// select a social space at uniform random from the agent's list of social spaces.
			// in reality this wouldn't be uniform random, rather mostly a function of proximity,
			// which can be implemented once geospatial attributes are added to spaces
			agent.setLocation(
				sim,
				agent.social_spaces[sampleUniform(0, int64(len(agent.social_spaces)-1))],
				sampleNormal(45*60*1000, 15*60*1000),
			)
		}
	case Office, SocialSpace, HealthCareSpace:
		agent.setLocation(
			sim,
			agent.household,
			sampleNormal(12*60*60*1000, 4*60*60*1000),
		)
	default:
		panic("this shouldn't happen")
	}
}

func (agent *Agent) setLocation(sim *Simulation, location *Space, duration float64) {
	previous_location_id := agent.location.id

	// remove agent from current location
	agent.location.removeAgent(sim, agent)

	// push agent to next location
	location.addAgent(sim, agent)

	// set the agent's location to next location
	agent.location = location
	agent.location_change_epoch = sim.epoch
	agent.next_move_epoch = sim.epoch + int64(math.Ceil(duration/float64(sim.time_step)))
	agent.dispatchLocationUpdateEvent(sim, previous_location_id)
}

func (agent *Agent) infect(sim *Simulation) {
	agent.infection_profile = sim.pathogen.generateInfectionProfile()
	agent.setState(sim, Infected)
}

func (agent *Agent) dispatchStateUpdateEvent(sim *Simulation, previous_state AgentState) {
	event := logger.Event{
		Type: AgentStateUpdate,
		Payload: AgentStateUpdatePayload{
			Epoch:               sim.epoch,
			Id:                  agent.id,
			State:               agent.state,
			PreviousState:       previous_state,
			HasInfectionProfile: agent.infection_profile != nil,

			jurisdiction: agent.household.jurisdiction,
		},
	}

	sim.logger.Log(event)
}

func (agent *Agent) dispatchLocationUpdateEvent(sim *Simulation, previous_location_id uuid.UUID) {
	event := logger.Event{
		Type: AgentLocationUpdate,
		Payload: AgentLocationUpdatePayload{
			Epoch:              sim.epoch,
			Id:                 agent.id,
			LocationId:         agent.location.id,
			PreviousLocationId: previous_location_id,
		},
	}

	sim.logger.Log(event)
}

func (agent *Agent) pInfected(sim *Simulation) float64 {
	_, volume, _, total_infectious_doses, policy := agent.location.state()

	filtration_efficiency := 0.0
	if policy != nil && policy.is_mask_mandate && agent.is_compliant {
		filtration_efficiency = agent.mask_filtration_efficiency
	}

	dose_concentration := total_infectious_doses / volume

	p := 1 - math.Exp(-1*(1-filtration_efficiency)*dose_concentration*(agent.pulmonary_ventilation_rate/3600)*(float64(sim.time_step)/1000))

	return p
}
