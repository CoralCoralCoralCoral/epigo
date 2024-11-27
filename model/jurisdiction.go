package model

import "github.com/CoralCoralCoralCoral/simulation-engine/geo"

type Jurisdiction struct {
	id     string
	parent *Jurisdiction
	policy *Policy
}

func newJurisdiction(id string, parent *Jurisdiction) *Jurisdiction {
	jur := Jurisdiction{
		id:     id,
		parent: parent,
	}

	return &jur
}

func jurisdictionsFromFeatures() []*Jurisdiction {
	features := geo.LoadFeatures()

	jurisdictions := make([]*Jurisdiction, 0, len(features))

	// create jurisdictions
	for _, feature := range features {
		jurisdictions = append(jurisdictions, newJurisdiction(feature.Code(), nil))
	}

	// assign parents
	for _, feature := range features {
		if parent_code := feature.ParentCode(); parent_code != "" {
			jur := findJurisdiction(jurisdictions, func(val *Jurisdiction) bool {
				return val.Id() == feature.Code()
			})

			// find parent_jurisdiction
			parent_jur := findJurisdiction(jurisdictions, func(val *Jurisdiction) bool {
				return val.Id() == parent_code
			})

			if jur != nil && parent_jur != nil {
				jur.parent = parent_jur
			}
		}
	}

	return jurisdictions
}

func sampleJurisdiction(jurisdictions []*Jurisdiction, msoa_sampler *geo.MSOASampler) *Jurisdiction {
	msoa := msoa_sampler.Sample()

	jur := findJurisdiction(jurisdictions, func(val *Jurisdiction) bool {
		return val.Id() == msoa.GISCode
	})

	return jur
}

func findJurisdiction(jurisdictions []*Jurisdiction, predicate func(value *Jurisdiction) bool) *Jurisdiction {
	for _, value := range jurisdictions {
		if predicate(value) {
			return value
		}
	}

	return nil
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
