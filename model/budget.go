package model

import (
	"slices"

	"github.com/CoralCoralCoralCoral/simulation-engine/logger"
)

type BudgetConfig struct {
	StartingBudget        float64
	TestCost              float64
	MaskCost              float64
	LockdownCostPerCapita float64
	GDPPerCapitaPerEpoch  float64
	TaxRate               float64
	DepartmentBudgetRate  float64

	BudgetUpdatePayload *BudgetUpdatePayload

	CostMultiplier   float64
	IncomeMultiplier float64

	sim    *Simulation
	logger *logger.Logger
}

func InitialiseBudget(sim *Simulation) BudgetConfig {
	// https://www.ons.gov.uk/employmentandlabourmarket/peopleinwork/earningsandworkinghours/timeseries/ybuy/lms
	config := BudgetConfig{
		StartingBudget:        1000000,
		TestCost:              59.99,
		MaskCost:              19.99,
		LockdownCostPerCapita: 2500.0,
		GDPPerCapitaPerEpoch:  (50000.0 / (48 * 38.5)) / (1000 * 60 * 60 / float64(sim.config.TimeStep)),
		TaxRate:               0.2,
		DepartmentBudgetRate:  0.025,
		CostMultiplier:        1.0,
		IncomeMultiplier:      1.0,
		sim:                   sim,
		logger:                &sim.logger,
	}

	config.BudgetUpdatePayload = &BudgetUpdatePayload{
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

				conf.spendBudget(float64(totalTests) * conf.TestCost)
			}
		case EpochEnd:
			if payload, ok := event.Payload.(EpochEndPayload); ok {
				if (payload.Epoch*payload.TimeStep)%(24*60*60*1000) != 0 {
					return
				}
				go conf.logger.Log(logger.Event{
					Type:    BudgetUpdate,
					Payload: conf.BudgetUpdatePayload,
				})
			}
		case AgentLocationUpdate:
			if payload, ok := event.Payload.(AgentLocationUpdatePayload); ok {
				if payload.agent.location.type_ == Office {
					conf.addBudget(float64(payload.agent.next_move_epoch-payload.Epoch) * conf.GDPPerCapitaPerEpoch * conf.TaxRate * conf.DepartmentBudgetRate)
				}
			}
		case CommandProcessed:
			if payload, ok := event.Payload.(CommandProcessedPayload); ok {
				if payload.Command.Type == ApplyPolicyUpdate {
					if command, ok := payload.Command.Payload.(*ApplyPolicyUpdatePayload); ok {
						conf.handleCommandProcessedPayload(command)
					}
				}
			}
		}
	}
}

func (conf *BudgetConfig) handleCommandProcessedPayload(payload *ApplyPolicyUpdatePayload) {
	var affectedPeople int
	var jur *Jurisdiction
	for _, _jur := range conf.sim.jurisdictions {
		if _jur.Id == payload.JurisdictionId {
			jur = _jur
		}
	}

	leafJurs := getLeafJuristictionIDs(jur)

	for _, office := range conf.sim.offices {
		if slices.Contains(leafJurs, &office.jurisdiction.Id) {
			affectedPeople += len(office.occupants)
		}
	}
	for _, houses := range conf.sim.households {
		if slices.Contains(leafJurs, &houses.jurisdiction.Id) {
			affectedPeople += len(houses.occupants)
		}
	}
	for _, social_space := range conf.sim.social_spaces {
		if slices.Contains(leafJurs, &social_space.jurisdiction.Id) {
			affectedPeople += len(social_space.occupants)
		}
	}

	if payload.IsLockdown != nil && *payload.IsLockdown {
		conf.spendBudget(float64(affectedPeople) * conf.LockdownCostPerCapita)
	}
	if payload.IsMaskMandate != nil && *payload.IsMaskMandate {
		conf.spendBudget(float64(affectedPeople) * conf.MaskCost)
	}
}

func (conf *BudgetConfig) spendBudget(amount float64) {
	conf.BudgetUpdatePayload.CurrentBudget -= amount * conf.CostMultiplier
}

func (conf *BudgetConfig) addBudget(amount float64) {
	conf.BudgetUpdatePayload.CurrentBudget += amount * conf.IncomeMultiplier
}

func getLeafJuristictionIDs(jur *Jurisdiction) []*string {
	children := jur.children
	temp := make([]*string, 1)
	if len(children) == 0 {
		temp = append(temp, &jur.Id)
	} else {
		for _, j := range children {
			temp = append(temp, getLeafJuristictionIDs(j)...)
		}
	}

	return temp
}
