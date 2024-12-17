package model

import "github.com/CoralCoralCoralCoral/simulation-engine/logger"

const BudgetUpdate logger.EventType = "budget_update"

type BudgetPayload struct {
	CurrentBudget float32 `json:"current_budget"`
}

type BudgetConfig struct {
	StartingBudget       float32
	TestCost             float32
	GDPPerCapitaPerEpoch float32
	TaxRate              float32
	DepartmentBudgetRate float32

	BudgetPayload *BudgetPayload

	CostMultiplier   float32
	IncomeMultiplier float32

	sim *logger.Logger
}

func InitialiseBudget(sim *Simulation) BudgetConfig {
	// https://www.ons.gov.uk/employmentandlabourmarket/peopleinwork/earningsandworkinghours/timeseries/ybuy/lms
	config := BudgetConfig{
		StartingBudget:       1000000,
		TestCost:             39.99,
		GDPPerCapitaPerEpoch: (50000.0 / (48 * 38.5)) / (1000 * 60 * 60 / float32(sim.config.TimeStep)),
		TaxRate:              0.2,
		DepartmentBudgetRate: 0.01,
		CostMultiplier:       1.0,
		IncomeMultiplier:     1.0,
		sim:                  &sim.logger,
	}

	config.BudgetPayload = &BudgetPayload{
		CurrentBudget: config.StartingBudget,
	}

	return config
}

func (conf *BudgetConfig) NewEventSubscriber() func(event *logger.Event) {

	return func(event *logger.Event) {
		switch event.Type {
		case SpaceTestingUpdate:
			if testPayload, ok := event.Payload.(SpaceTestingUpdatePayload); ok {
				totalTests := testPayload.Positives + testPayload.Negatives

				conf.BudgetPayload.CurrentBudget -= float32(totalTests) * conf.TestCost * conf.CostMultiplier
			}
		case EpochEnd:
			if payload, ok := event.Payload.(EpochEndPayload); ok {
				if (payload.Epoch*payload.TimeStep)%(24*60*60*1000) != 0 {
					return
				}
				go conf.sim.Log(logger.Event{
					Type:    BudgetUpdate,
					Payload: conf.BudgetPayload,
				})
			}
		case AgentLocationUpdate:
			if payload, ok := event.Payload.(AgentLocationUpdatePayload); ok {
				if payload.agent.location.type_ == Office {
					conf.BudgetPayload.CurrentBudget += float32(payload.agent.next_move_epoch-payload.Epoch) * conf.GDPPerCapitaPerEpoch * conf.TaxRate * conf.DepartmentBudgetRate * conf.IncomeMultiplier
				}
			}
		}
	}
}
