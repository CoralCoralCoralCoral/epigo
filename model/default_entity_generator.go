package model

import "github.com/CoralCoralCoralCoral/simulation-engine/geo"

type DefaultEntityGenerator struct {
}

func NewDefaultEntityGenerator() *DefaultEntityGenerator {
	return &DefaultEntityGenerator{}
}

func (g *DefaultEntityGenerator) Generate() Entities {
	msoa_sampler := geo.NewMSOASampler()

	return Entities{}
}
