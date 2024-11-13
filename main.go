package main

import (
	"bufio"
	"fmt"
	"math"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"gonum.org/v1/gonum/stat/distuv"
)

const Susceptible = "susceptible"
const Infected = "infected"
const Infectious = "infectious"
const Immune = "immune"

const Household = "household"
const Office = "office"
const SocialSpace = "social_space"

type State struct {
	id              uuid.UUID
	pathogen        PathogenProfile
	start_time      time.Time
	epoch           int64
	time_step       int64
	agents          []*Agent
	households      []*Space
	offices         []*Space
	social_spaces   []*Space
	is_mask_mandate bool
	is_lockdown     bool
}

type PathogenProfile struct {
	incubation_period    [2]float64
	recovery_period      [2]float64
	immunity_period      [2]float64
	quanta_emission_rate [2]float64
}

type Agent struct {
	id                           uuid.UUID
	household                    *Space
	office                       *Space
	social_spaces                []*Space
	location                     *Space
	location_change_epoch        int64
	infection_state              string
	infection_state_change_epoch int64
	infectious_contacts          []InfectiousContact
	pulmonary_ventilation_rate   float64
	is_compliant                 bool
}

type InfectiousContact struct {
	id               uuid.UUID
	infectious_epoch int64
}

type Space struct {
	mu                     *sync.RWMutex
	id                     uuid.UUID
	type_                  string
	occupants              []*Agent
	capacity               int64
	volume                 float64
	air_change_rate        float64
	total_infectious_doses float64
}

func main() {
	// a covid-like pathogen
	pathogen := PathogenProfile{
		incubation_period:    [2]float64{3 * 24 * 60 * 60 * 1000, 8 * 60 * 60 * 1000},
		recovery_period:      [2]float64{7 * 24 * 60 * 60 * 1000, 8 * 60 * 60 * 1000},
		immunity_period:      [2]float64{330 * 24 * 60 * 60 * 1000, 90 * 24 * 60 * 60 * 1000},
		quanta_emission_rate: [2]float64{100, 20},
	}

	// create a game with 150k people
	state := createState(150000, pathogen)

	state.start()
}

func (state *State) start() {
	ln, err := net.Listen("tcp", ":9876")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer ln.Close()

	fmt.Println("Waiting for player to connect on port 9876")

	conn, err := ln.Accept()
	if err != nil {
		fmt.Println("Error accepting connection:", err)
		return
	}
	defer conn.Close()

	commands := make(chan string)

	fmt.Println("Player connected")
	fmt.Println("Starting game...")

	metrics_engine := newMetricsEngine(state, 24*60*60*1000)
	state.infect_random_agent()

	go func() {
		for {
			command, _ := bufio.NewReader(conn).ReadString('\n')
			commands <- command
		}
	}()

	for {
		select {
		case command := <-commands:
			if command == "mask mandate\n" {
				state.toggle_mask_mandate()
			}

			if command == "lockdown\n" {
				state.toggle_lockdown()
			}
		default:
			state.simulate_epoch()
			metrics_engine.report_metrics()
		}
	}
}

func (state *State) simulate_epoch() {
	for _, agent := range state.agents {
		agent.update(state)
		agent.move(state)
	}

	for _, household := range state.households {
		household.update(state)
	}

	for _, office := range state.offices {
		office.update(state)
	}

	for _, social_space := range state.social_spaces {
		social_space.update(state)
	}

	state.epoch = state.epoch + 1
}

func (state *State) infect_random_agent() {
	agent_idx := sampleUniform(0, int64(len(state.agents)-1))
	state.agents[agent_idx].infection_state = Infected
}

func (state *State) toggle_mask_mandate() {
	state.is_mask_mandate = !state.is_mask_mandate
}

func (state *State) toggle_lockdown() {
	state.is_lockdown = !state.is_lockdown
}

func (state *State) time() time.Time {
	return state.start_time.Add(time.Duration(state.epoch*state.time_step) * time.Millisecond)
}

func (space *Space) update(state *State) {
	space.mu.Lock()
	defer space.mu.Unlock()

	for _, occupant := range space.occupants {
		if occupant.infection_state == Infectious {
			filtration_efficiency := 0.0
			if state.is_mask_mandate && occupant.is_compliant {
				filtration_efficiency = math.Max(sampleNormal(0.95, 0.15), 1)
			}

			quanta_emission_rate := (1 - filtration_efficiency) * sampleNormal(state.pathogen.quanta_emission_rate[0], state.pathogen.quanta_emission_rate[1]) / 3600

			space.total_infectious_doses += quanta_emission_rate * float64(state.time_step) / 1000
		}
	}

	ventilation_rate := space.volume * space.air_change_rate / 3600
	space.total_infectious_doses = space.total_infectious_doses * (1 / math.Exp((ventilation_rate/space.volume)*float64(state.time_step)/1000))
}

func (space *Space) removeAgent(agent *Agent) {
	space.mu.Lock()
	defer space.mu.Unlock()

	for idx, candidate := range space.occupants {
		if candidate.id == agent.id {
			space.occupants = append(space.occupants[:idx], space.occupants[idx+1:]...)
			break
		}
	}
}

func (space *Space) addAgent(agent *Agent) {
	space.mu.Lock()
	defer space.mu.Unlock()

	space.occupants = append(space.occupants, agent)
}

func (space *Space) state() (float64, float64, float64) {
	space.mu.RLock()
	defer space.mu.RUnlock()

	return space.volume, space.air_change_rate, space.total_infectious_doses
}

func (space *Space) infectiousContacts() []InfectiousContact {
	snapshot := make([]InfectiousContact, 0)
	for _, occupant := range space.occupants {
		if occupant.infection_state == Infectious {
			snapshot = append(snapshot, InfectiousContact{id: occupant.id, infectious_epoch: occupant.infection_state_change_epoch})
		}
	}

	return snapshot
}

func (agent *Agent) move(state *State) {
	location_duration := float64((state.epoch - agent.location_change_epoch) * state.time_step)

	var next_location *Space = nil

	switch agent.location.type_ {
	case Household:
		if state.is_lockdown && agent.is_compliant {
			break
		}

		sample_duration := sampleNormal(12*60*60*1000, 4*60*60*1000)

		if location_duration >= sample_duration {
			p_goes_to_office := 0.55

			goes_to_office := sampleBernoulli(p_goes_to_office)

			if goes_to_office == 1 {
				next_location = agent.office
			} else {
				// select a social space at uniform random from the agent's list of social spaces.
				// in reality this wouldn't be uniform random, rather mostly a function of proximity,
				// which can be implemented once geospatial attributes are added to spaces
				next_location = agent.social_spaces[sampleUniform(0, int64(len(agent.social_spaces)-1))]
			}
		}
	case Office:
		sample_duration := sampleNormal(8*60*60*1000, 2*60*60*1000)

		if location_duration >= sample_duration {
			next_location = agent.household
		}
	case SocialSpace:
		sample_duration := sampleNormal(45*60*1000, 15*60*1000)

		if location_duration >= sample_duration {
			next_location = agent.household
		}
	default:
		panic("this shouldn't happen")
	}

	if next_location != nil {
		// remove agent from current location
		agent.location.removeAgent(agent)

		// push agent to next location's list of occupants
		next_location.addAgent(agent)

		// set the agent's location to next location
		agent.location = next_location
		agent.location_change_epoch = state.epoch
	}
}

func (agent *Agent) update(state *State) {
	infection_state_duration := float64((state.epoch - agent.infection_state_change_epoch) * state.time_step)

	switch agent.infection_state {
	case Susceptible:
		is_infected := sampleBernoulli(agent.pInfected(state))

		if is_infected == 1 {
			agent.infection_state = Infected
			agent.infection_state_change_epoch = state.epoch
			agent.infectious_contacts = agent.location.infectiousContacts()

		}
	case Infected:
		incubation_period := sampleNormal(state.pathogen.incubation_period[0], state.pathogen.incubation_period[1])

		if infection_state_duration >= incubation_period {
			agent.infection_state = Infectious
			agent.infection_state_change_epoch = state.epoch
		}
	case Infectious:
		recovery_period := sampleNormal(state.pathogen.recovery_period[0], state.pathogen.recovery_period[1])

		if infection_state_duration >= recovery_period {
			agent.infection_state = Immune
			agent.infection_state_change_epoch = state.epoch
		}
	case Immune:
		immunity_period := sampleNormal(state.pathogen.immunity_period[0], state.pathogen.immunity_period[1])

		if infection_state_duration >= immunity_period {
			agent.infection_state = Susceptible
			agent.infection_state_change_epoch = state.epoch
		}
	default:
		panic("this shouldn't be possible")
	}
}

func (agent *Agent) pInfected(state *State) float64 {
	volume, air_change_rate, total_infectious_doses := agent.location.state()
	ventilationRate := volume * air_change_rate / 3600

	filtration_efficiency := 0.0
	if state.is_mask_mandate && agent.is_compliant {
		filtration_efficiency = math.Max(sampleNormal(0.95, 0.15), 1)
	}

	return 1 - 1/math.Exp(((1-filtration_efficiency)*total_infectious_doses*agent.pulmonary_ventilation_rate/3600*float64(state.time_step)/1000)/ventilationRate)
}

func createState(agent_count int64, pathogen_profile PathogenProfile) State {
	households := createHouseholds(agent_count)
	offices := createOffices(agent_count)
	social_spaces := createSocialSpaces(agent_count)
	agents := createAgents(agent_count, households, offices, social_spaces)

	return State{
		id:            uuid.New(),
		pathogen:      pathogen_profile,
		start_time:    time.Now(),
		epoch:         0,
		time_step:     15 * 60 * 1000,
		agents:        agents,
		households:    households,
		offices:       offices,
		social_spaces: social_spaces,
	}
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
		agent := newAgent(nil, nil)

		num_social_spaces := int(math.Max(1, math.Floor(sampleNormal(5, 4))))
		for i := 0; i < num_social_spaces; i++ {
			agent.social_spaces = append(agent.social_spaces, social_spaces[sampleUniform(0, int64(len(social_spaces)-1))])
		}

		agents[i] = &agent
	}

	// allocate agents to households
	householdIdx, householdAllocatedCapacity := 0, 0
	for _, agent := range agents {
		household := households[householdIdx]
		agent.household = household
		agent.location = household

		householdAllocatedCapacity += 1
		if householdAllocatedCapacity == int(household.capacity) {
			householdIdx += 1
			householdAllocatedCapacity = 0
		}
	}

	// allocate agents to offices
	officeIdx, officeAllocatedCapacity := 0, 0
	for _, agent := range agents {
		office := offices[officeIdx]
		agent.office = office

		officeAllocatedCapacity += 1
		if officeAllocatedCapacity == int(office.capacity) {
			officeIdx += 1
			officeAllocatedCapacity = 0
		}
	}

	return agents
}

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

func newAgent(household, office *Space) Agent {
	is_compliant := false
	if sampleBernoulli(0.95) == 1 {
		is_compliant = true
	}

	return Agent{
		id:                           uuid.New(),
		household:                    household,
		office:                       office,
		social_spaces:                make([]*Space, 0),
		location:                     household,
		location_change_epoch:        0,
		infection_state:              Susceptible,
		infection_state_change_epoch: 0,
		infectious_contacts:          make([]InfectiousContact, 0),
		pulmonary_ventilation_rate:   sampleNormal(0.36, 0.01),
		is_compliant:                 is_compliant,
	}
}

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
