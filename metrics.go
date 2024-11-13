package main

import (
	"fmt"
	"math"
)

type MetricsEngine struct {
	state              *State
	reporting_interval int64
	reporting_period   int
}

type Metrics struct {
	infections      int64
	infectious      int64
	incidences      int64
	serial_interval SerialInterval
}

// SerialInterval represents the serial interval with its mean and standard deviation
type SerialInterval struct {
	mean float64
	sd   float64
}

func (engine *MetricsEngine) report_metrics() {
	if (engine.state.epoch*engine.state.time_step)%engine.reporting_interval != 0 {
		return
	}

	period_start_epoch := engine.state.epoch - engine.reporting_interval/engine.state.time_step

	metrics := engine.generate_metrics(period_start_epoch)

	fmt.Print("\033[H\033[2J")

	fmt.Printf("Epidemic state on %s\n", engine.state.time().Format("02-01-2006"))
	fmt.Printf("	New infections:			%d\n", metrics.incidences)
	fmt.Printf("	Active infections:		%d\n", metrics.infections)
	fmt.Printf("	Mean serial interval:		%0.2f\n", metrics.serial_interval.mean)

	interventions := "none"
	if engine.state.is_mask_mandate && engine.state.is_lockdown {
		interventions = "mask mandate, lockdown"
	} else if engine.state.is_mask_mandate {
		interventions = "mask mandate"
	} else if engine.state.is_lockdown {
		interventions = "lockdown"
	}

	fmt.Printf("	Interventions in effect:	%s\n", interventions)
}

func (engine *MetricsEngine) generate_metrics(period_start_epoch int64) Metrics {
	metrics := Metrics{
		infections: 0,
		infectious: 0,
		incidences: 0,
		serial_interval: SerialInterval{
			mean: 0,
			sd:   0,
		},
	}

	serial_intervals := make([]float64, 0)

	for _, agent := range engine.state.agents {
		if agent.infection_state == Infected || agent.infection_state == Infectious {
			metrics.infections += 1
		}

		if agent.infection_state == Infected && agent.infection_state_change_epoch >= period_start_epoch {
			metrics.incidences += 1
		}

		if agent.infection_state == Infectious {
			metrics.infectious += 1

			if len(agent.infectious_contacts) > 0 {
				cumulative_serial_interval := 0.0
				for _, contact := range agent.infectious_contacts {
					cumulative_serial_interval += ((float64(agent.infection_state_change_epoch) - float64(contact.infectious_epoch)) * float64(engine.state.time_step)) / (24 * 60 * 60 * 1000)
				}

				serial_intervals = append(serial_intervals, cumulative_serial_interval/float64(len(agent.infectious_contacts)))
			}
		}
	}

	metrics.serial_interval.mean = calculateMean(serial_intervals)
	metrics.serial_interval.sd = calculateStandardDeviation(serial_intervals)

	engine.reporting_period++
	return metrics
}

func newMetricsEngine(state *State, reporting_interval int64) MetricsEngine {
	return MetricsEngine{
		state:              state,
		reporting_interval: reporting_interval,
		reporting_period:   0,
	}
}

// Calculate the mean of a slice of float64
func calculateMean(data []float64) float64 {
	var sum float64
	for _, value := range data {
		sum += value
	}
	return sum / float64(len(data))
}

// Calculate the standard deviation of a slice of float64
func calculateStandardDeviation(data []float64) float64 {
	mean := calculateMean(data)
	var variance float64
	for _, value := range data {
		variance += math.Pow(value-mean, 2)
	}
	variance /= float64(len(data)) // Population SD, use len(data)-1 for sample SD
	return math.Sqrt(variance)
}
