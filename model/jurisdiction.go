package model

import (
	"encoding/json"

	"github.com/CoralCoralCoralCoral/simulation-engine/geo"
	"github.com/CoralCoralCoralCoral/simulation-engine/logger"
)

type Jurisdiction struct {
	Id      string       `json:"id"`
	Policy  *Policy      `json:"policy"`
	Feature *geo.Feature `json:"feature"`

	// non serialized fields
	parent   *Jurisdiction
	children []*Jurisdiction
}

func (jur *Jurisdiction) Parent() *Jurisdiction {
	return jur.parent
}

func newJurisdiction(config *Config, id string, feature *geo.Feature) *Jurisdiction {
	jur := Jurisdiction{
		Id:       id,
		children: make([]*Jurisdiction, 0),
		Policy: &Policy{
			TestStrategy:           TestNone,
			TestCapacityMultiplier: 1,
			ComplianceProbability:  config.ComplianceProbability,
		},
		Feature: feature,
	}

	return &jur
}

// func (jur *Jurisdiction) Id() string {
// 	return jur.id
// }

// make assigning parent an explicit operation
func (jur *Jurisdiction) assignParent(parent *Jurisdiction) {
	jur.parent = parent

	parent.children = append(parent.children, jur)
}

// func (jur *Jurisdiction) applyPolicy(policy *Policy) {
// 	jur.policy = policy
// }

func (jur *Jurisdiction) applyPolicyUpdate(sim *Simulation, update *ApplyPolicyUpdatePayload) {
	if update.IsLockdown != nil {
		jur.Policy.IsLockdown = *update.IsLockdown
	}

	if update.IsMaskMandate != nil {
		jur.Policy.IsMaskMandate = *update.IsMaskMandate
	}

	if update.IsSelfIsolationMandate != nil {
		jur.Policy.IsSelfIsolationMandate = *update.IsSelfIsolationMandate
	}

	if update.IsSelfReportingMandate != nil {
		jur.Policy.IsSelfReportingMandate = *update.IsSelfReportingMandate
	}

	if update.TestCapacityMultiplier != nil {
		jur.Policy.TestCapacityMultiplier = *update.TestCapacityMultiplier
	}

	if update.TestStrategy != nil {
		jur.Policy.TestStrategy = *update.TestStrategy
	}

	if update.ComplianceProbability != nil {
		jur.Policy.ComplianceProbability = *update.ComplianceProbability
	}

	var updatedPolicy Policy
	policyBytes, _ := json.Marshal(jur.Policy)
	json.Unmarshal(policyBytes, &updatedPolicy)

	sim.logger.Log(logger.Event{
		Type: PolicyUpdate,
		Payload: PolicyUpdatePayload{
			JurisdictionId: jur.Id,
			Policy:         updatedPolicy,
		},
	})

	// recursively apply policy update to sub jurisdictions
	for _, child := range jur.children {
		child.applyPolicyUpdate(sim, update)
	}
}

func (jur *Jurisdiction) resolvePolicy() (policy *Policy) {
	// if jur.parent != nil {
	// 	policy = jur.parent.resolvePolicy()
	// }

	// if policy == nil {
	// 	policy = jur.policy
	// }

	return jur.Policy
}
