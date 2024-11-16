package model

type Pathogen struct {
	IncubationPeriod   [2]float64
	RecoveryPeriod     [2]float64
	ImmunityPeriod     [2]float64
	QuantaEmissionRate [2]float64
}

type InfectionProfile struct {
	incubation_period    float64
	recovery_period      float64
	immunity_period      float64
	quanta_emission_rate float64
}

func (pathogen *Pathogen) generateInfectionProfile() *InfectionProfile {
	return &InfectionProfile{
		incubation_period:    sampleNormal(pathogen.IncubationPeriod[0], pathogen.IncubationPeriod[1]),
		recovery_period:      sampleNormal(pathogen.RecoveryPeriod[0], pathogen.RecoveryPeriod[1]),
		immunity_period:      sampleNormal(pathogen.ImmunityPeriod[0], pathogen.ImmunityPeriod[1]),
		quanta_emission_rate: sampleNormal(pathogen.QuantaEmissionRate[0], pathogen.QuantaEmissionRate[1]),
	}
}
