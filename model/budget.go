package model

import (
	"slices"

	"github.com/CoralCoralCoralCoral/simulation-engine/logger"
)

const BudgetUpdate logger.EventType = "budget_update"

type BudgetPayload struct {
	CurrentBudget float32 `json:"current_budget"`
}

type BudgetConfig struct {
	StartingBudget        float32
	TestCost              float32
	MaskCost              float32
	LockdownCostPerCapita float32
	GDPPerCapitaPerEpoch  float32
	TaxRate               float32
	DepartmentBudgetRate  float32

	BudgetPayload *BudgetPayload

	CostMultiplier   float32
	IncomeMultiplier float32

	sim    *Simulation
	logger *logger.Logger
}

func InitialiseBudget(sim *Simulation) BudgetConfig {
	// https://www.ons.gov.uk/employmentandlabourmarket/peopleinwork/earningsandworkinghours/timeseries/ybuy/lms
	config := BudgetConfig{
		StartingBudget:        1000000,
		TestCost:              39.99,
		MaskCost:              9.99,
		LockdownCostPerCapita: 1000.0,
		GDPPerCapitaPerEpoch:  (50000.0 / (48 * 38.5)) / (1000 * 60 * 60 / float32(sim.config.TimeStep)),
		TaxRate:               0.2,
		DepartmentBudgetRate:  0.01,
		CostMultiplier:        1.0,
		IncomeMultiplier:      1.0,
		sim:                   sim,
		logger:                &sim.logger,
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

				conf.spendBudget(float32(totalTests) * conf.TestCost)
			}
		case EpochEnd:
			if payload, ok := event.Payload.(EpochEndPayload); ok {
				if (payload.Epoch*payload.TimeStep)%(24*60*60*1000) != 0 {
					return
				}
				go conf.logger.Log(logger.Event{
					Type:    BudgetUpdate,
					Payload: conf.BudgetPayload,
				})
			}
		case AgentLocationUpdate:
			if payload, ok := event.Payload.(AgentLocationUpdatePayload); ok {
				if payload.agent.location.type_ == Office {
					conf.addBudget(float32(payload.agent.next_move_epoch-payload.Epoch) * conf.GDPPerCapitaPerEpoch * conf.TaxRate * conf.DepartmentBudgetRate)
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
		if _jur.Id() == payload.JurisdictionId {
			jur = _jur
		}
	}

	leafJurs := getLeafJuristictionIDs(jur)

	for _, office := range conf.sim.offices {
		if slices.Contains(leafJurs, &office.jurisdiction.id) {
			affectedPeople += len(office.occupants)
		}
	}
	for _, houses := range conf.sim.households {
		if slices.Contains(leafJurs, &houses.jurisdiction.id) {
			affectedPeople += len(houses.occupants)
		}
	}
	for _, social_space := range conf.sim.social_spaces {
		if slices.Contains(leafJurs, &social_space.jurisdiction.id) {
			affectedPeople += len(social_space.occupants)
		}
	}
	println("Affected People: %d", affectedPeople)

	if payload.IsLockdown != nil && *payload.IsLockdown {
		conf.spendBudget(float32(affectedPeople) * conf.LockdownCostPerCapita)
	}
	if payload.IsMaskMandate != nil && *payload.IsMaskMandate {
		conf.spendBudget(float32(affectedPeople) * conf.MaskCost)
	}
}

func (conf *BudgetConfig) spendBudget(amount float32) {
	conf.BudgetPayload.CurrentBudget -= amount * conf.CostMultiplier
}

func (conf *BudgetConfig) addBudget(amount float32) {
	conf.BudgetPayload.CurrentBudget += amount * conf.IncomeMultiplier
}

func getLeafJuristictionIDs(jur *Jurisdiction) []*string {
	children := jur.children
	temp := make([]*string, 1)
	if children == nil || len(children) == 0 {
		temp = append(temp, &jur.id)
	} else {
		for _, j := range children {
			temp = append(temp, getLeafJuristictionIDs(j)...)
		}
	}

	return temp
}
