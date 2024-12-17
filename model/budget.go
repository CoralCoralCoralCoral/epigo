package model

import "github.com/CoralCoralCoralCoral/simulation-engine/logger"

const BudgetUpdate logger.EventType = "budget_update"

type BudgetPayload struct {
	CurrentBudget float32 `json:"current_budget"`
}

type BudgetConfig struct {
	StartingBudget float32
	TestCost       float32

	BudgetPayload *BudgetPayload

	CostMultiplier float32
	sim            *logger.Logger
}

func InitialiseBudget(sim *logger.Logger) BudgetConfig {
	config := BudgetConfig{
		StartingBudget: 1000000,
		TestCost:       39.99,
		CostMultiplier: 1.0,
		sim:            sim,
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

				conf.BudgetPayload.CurrentBudget -= float32(totalTests) * conf.TestCost
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
		}
	}
}
