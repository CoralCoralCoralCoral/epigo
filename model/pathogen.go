package model

type Pathogen struct {
	IncubationPeriod           [2]float64 `json:"incubation_period"`
	RecoveryPeriod             [2]float64 `json:"recovery_period"`
	ImmunityPeriod             [2]float64 `json:"immunity_period"`
	PrehospitalizationPeriod   [2]float64 `json:"prehospitalization_period"`
	HospitalizationPeriod      [2]float64 `json:"hospitalization_period"`
	QuantaEmissionRate         [2]float64 `json:"quanta_emission_rate"`
	HospitalizationProbability float64    `json:"hospitalization_probability"`
	DeathProbability           float64    `json:"death_probability"` // conditional on hospitalized
	AsymptomaticProbability    float64    `json:"asymptomatic_probability"`
}

type InfectionProfile struct {
	incubation_period         float64
	recovery_period           float64
	immunity_period           float64
	prehospitalization_period float64
	hospitalization_period    float64
	quanta_emission_rate      float64
	is_hospitalized           bool
	is_dead                   bool // conditional on hospitalized
	is_asymptomatic           bool
}

func (pathogen *Pathogen) generateInfectionProfile() *InfectionProfile {
	is_hospitalized := false
	if sampleBernoulli(pathogen.HospitalizationProbability) == 1 {
		is_hospitalized = true
	}

	is_dead := false
	if is_hospitalized && sampleBernoulli(pathogen.DeathProbability) == 1 {
		is_dead = true
	}

	prehospitalization_period := 0.0
	hospitalization_period := 0.0
	if is_hospitalized {
		prehospitalization_period = sampleNormal(pathogen.PrehospitalizationPeriod[0], pathogen.PrehospitalizationPeriod[1])
		hospitalization_period = sampleNormal(pathogen.HospitalizationPeriod[0], pathogen.HospitalizationPeriod[1])
	}

	is_asymptomatic := false
	if sampleBernoulli(pathogen.AsymptomaticProbability) == 1 {
		is_asymptomatic = true
	}

	return &InfectionProfile{
		incubation_period:         sampleNormal(pathogen.IncubationPeriod[0], pathogen.IncubationPeriod[1]),
		recovery_period:           sampleNormal(pathogen.RecoveryPeriod[0], pathogen.RecoveryPeriod[1]),
		immunity_period:           sampleNormal(pathogen.ImmunityPeriod[0], pathogen.ImmunityPeriod[1]),
		prehospitalization_period: prehospitalization_period,
		hospitalization_period:    hospitalization_period,
		quanta_emission_rate:      sampleNormal(pathogen.QuantaEmissionRate[0], pathogen.QuantaEmissionRate[1]),
		is_hospitalized:           is_hospitalized,
		is_dead:                   is_dead,
		is_asymptomatic:           is_asymptomatic,
	}
}
