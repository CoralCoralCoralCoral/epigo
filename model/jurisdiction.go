package model

import "github.com/CoralCoralCoralCoral/simulation-engine/geo"

type Jurisdiction struct {
	id      string
	parent  *Jurisdiction
	policy  *Policy
	feature *geo.Feature
}

func (jur *Jurisdiction) Parent() *Jurisdiction {
	return jur.parent
}

func newJurisdiction(id string, parent *Jurisdiction, feature *geo.Feature) *Jurisdiction {
	jur := Jurisdiction{
		id:     id,
		parent: parent,
		// temporarily use a default test policy of test everyone
		policy: &Policy{
			TestStrategy: TestEveryone,
		},
		feature: feature,
	}

	return &jur
}

func (jur *Jurisdiction) Id() string {
	return jur.id
}

func (jur *Jurisdiction) applyPolicy(policy *Policy) {
	jur.policy = policy
}

func (jur *Jurisdiction) resolvePolicy() (policy *Policy) {
	if jur.parent != nil {
		policy = jur.parent.resolvePolicy()
	}

	if policy == nil {
		policy = jur.policy
	}

	return
}
