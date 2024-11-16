package model

import (
	"fmt"
	"sync"

	"github.com/umran/epigo/logger"
)

type Metrics struct {
	mu                    *sync.RWMutex
	new_infections        int
	new_recoveries        int
	infected_population   int
	infectious_population int
	immune_population     int
	is_mask_mandate       bool
	is_lockdown           bool
}

func newMetrics() *Metrics {
	return &Metrics{
		mu: new(sync.RWMutex),
	}
}

func (metrics *Metrics) applyEvent(event *logger.Event) {
	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	switch event.Type {
	case AgentStateUpdate:
		if payload, ok := event.Payload.(AgentStateUpdatePayload); ok {
			switch payload.State {
			case Infected:
				metrics.new_infections += 1
				metrics.infected_population += 1
			case Infectious:
				metrics.infectious_population += 1
			case Immune:
				metrics.immune_population += 1
				metrics.new_recoveries += 1
				metrics.infected_population -= 1
				metrics.infectious_population -= 1
			case Susceptible:
				metrics.immune_population -= 1
			default:
				panic("this should not be possible")
			}
		}
	case CommandProcessed:
		if payload, ok := event.Payload.(CommandProcessedPayload); ok {
			if payload.Command == "mask mandate\n" {
				metrics.is_mask_mandate = !metrics.is_mask_mandate
			}

			if payload.Command == "lockdown\n" {
				metrics.is_lockdown = !metrics.is_lockdown
			}
		}
	default:
		// ignore other types of events
	}
}

func (metrics *Metrics) reset() {
	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	metrics.new_infections = 0
	metrics.new_recoveries = 0
}

func (metrics *Metrics) print(date string) {
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()

	fmt.Print("\033[H\033[2J")

	fmt.Printf("Epidemic state on %s\n", date)
	fmt.Printf("	New infections:			%d\n", metrics.new_infections)
	fmt.Printf("	New recoveries:			%d\n", metrics.new_recoveries)
	fmt.Printf("	Infected population:		%d\n", metrics.infected_population)
	fmt.Printf("	Infectious population:		%d\n", metrics.infectious_population)
	fmt.Printf("	Immune population:		%d\n", metrics.immune_population)

	interventions := "none"
	if metrics.is_mask_mandate && metrics.is_lockdown {
		interventions = "mask mandate, lockdown"
	} else if metrics.is_mask_mandate {
		interventions = "mask mandate"
	} else if metrics.is_lockdown {
		interventions = "lockdown"
	}

	fmt.Printf("	Interventions in effect:	%s\n", interventions)
}

func NewMetricsSubscriber() func(event *logger.Event) {
	metrics := newMetrics()

	return func(event *logger.Event) {
		switch event.Type {
		case EpochEnd:
			if payload, ok := event.Payload.(EpochEndPayload); ok {
				if (payload.Epoch*payload.TimeStep)%(24*60*60*1000) != 0 {
					return
				}
				metrics.print(payload.Time.Format("02-01-2006"))
				metrics.reset()
			} else {
				panic("unexpected event payload")
			}
		default:
			metrics.applyEvent(event)
		}
	}
}
