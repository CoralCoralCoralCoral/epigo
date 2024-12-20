package model

import (
	"github.com/CoralCoralCoralCoral/simulation-engine/geo"
	"github.com/CoralCoralCoralCoral/simulation-engine/logger"
)

type Jurisdiction struct {
	id       string
	parent   *Jurisdiction
	children []*Jurisdiction
	policy   *Policy
	feature  *geo.Feature
}

func (jur *Jurisdiction) Parent() *Jurisdiction {
	return jur.parent
}

func newJurisdiction(id string, feature *geo.Feature) *Jurisdiction {
	jur := Jurisdiction{
		id:       id,
		children: make([]*Jurisdiction, 0),
		policy: &Policy{
			test_strategy: TestNone,
		},
		feature: feature,
	}

	return &jur
}

func (jur *Jurisdiction) Id() string {
	return jur.id
}

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
		jur.policy.is_lockdown = *update.IsLockdown
	}

	if update.IsMaskMandate != nil {
		jur.policy.is_mask_mandate = *update.IsMaskMandate
	}

	if update.TestCapacityMultiplier != nil {
		jur.policy.test_capacity_multiplier = *update.TestCapacityMultiplier
	}

	if update.TestStrategy != nil {
		jur.policy.test_strategy = *update.TestStrategy
	}

	sim.logger.Log(logger.Event{
		Type: PolicyUpdate,
		Payload: PolicyUpdatePayload{
			JurisdictionId:         jur.id,
			IsMaskMandate:          jur.policy.is_mask_mandate,
			IsLockdown:             jur.policy.is_lockdown,
			TestStrategy:           jur.policy.test_strategy,
			TestCapacityMultiplier: jur.policy.test_capacity_multiplier,
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

	return jur.policy
}
