package model

type EntityGenerator interface {
	Generate() Entities
}

type Entities struct {
	agents            []*Agent
	jurisdictions     []*Jurisdiction
	households        []*Space
	offices           []*Space
	social_spaces     []*Space
	healthcare_spaces []*Space
}
