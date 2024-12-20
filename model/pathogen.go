package model

type Pathogen struct {
	incubation_period_mean         float64
	incubation_period_sd           float64
	recovery_period_mean           float64
	recovery_period_sd             float64
	immunity_period_mean           float64
	immunity_period_sd             float64
	prehospitalization_period_mean float64
	prehospitalization_period_sd   float64
	hospitalization_period_mean    float64
	hospitalization_period_sd      float64
	quanta_emission_rate_mean      float64
	quanta_emission_rate_sd        float64
	hospitalization_probability    float64
	death_probability              float64 // conditional on hospitalized
	asymptomatic_probability       float64
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

func newPathogen(config *Config) *Pathogen {
	return &Pathogen{
		incubation_period_mean:         config.IncubationPeriodMean,
		incubation_period_sd:           config.IncubationPeriodSd,
		recovery_period_mean:           config.RecoveryPeriodMean,
		recovery_period_sd:             config.RecoveryPeriodSd,
		immunity_period_mean:           config.ImmunityPeriodMean,
		immunity_period_sd:             config.ImmunityPeriodSd,
		prehospitalization_period_mean: config.PrehospitalizationPeriodMean,
		prehospitalization_period_sd:   config.PrehospitalizationPeriodSd,
		hospitalization_period_mean:    config.HospitalizationPeriodMean,
		hospitalization_period_sd:      config.HospitalizationPeriodSd,
		quanta_emission_rate_mean:      config.QuantaEmissionRateMean,
		quanta_emission_rate_sd:        config.QuantaEmissionRateSd,
		hospitalization_probability:    config.HospitalizationProbability,
		death_probability:              config.DeathProbability,
		asymptomatic_probability:       config.AsymptomaticProbability,
	}
}

func (pathogen *Pathogen) generateInfectionProfile() *InfectionProfile {
	is_hospitalized := false
	if sampleBernoulli(pathogen.hospitalization_probability) == 1 {
		is_hospitalized = true
	}

	is_dead := false
	if is_hospitalized && sampleBernoulli(pathogen.death_probability) == 1 {
		is_dead = true
	}

	prehospitalization_period := 0.0
	hospitalization_period := 0.0
	if is_hospitalized {
		prehospitalization_period = sampleNormal(pathogen.prehospitalization_period_mean, pathogen.prehospitalization_period_sd)
		hospitalization_period = sampleNormal(pathogen.hospitalization_period_mean, pathogen.hospitalization_period_sd)
	}

	is_asymptomatic := false
	if sampleBernoulli(pathogen.asymptomatic_probability) == 1 {
		is_asymptomatic = true
	}

	return &InfectionProfile{
		incubation_period:         sampleNormal(pathogen.incubation_period_mean, pathogen.incubation_period_sd),
		recovery_period:           sampleNormal(pathogen.recovery_period_mean, pathogen.recovery_period_sd),
		immunity_period:           sampleNormal(pathogen.immunity_period_mean, pathogen.immunity_period_sd),
		prehospitalization_period: prehospitalization_period,
		hospitalization_period:    hospitalization_period,
		quanta_emission_rate:      sampleNormal(pathogen.quanta_emission_rate_mean, pathogen.quanta_emission_rate_sd),
		is_hospitalized:           is_hospitalized,
		is_dead:                   is_dead,
		is_asymptomatic:           is_asymptomatic,
	}
}
