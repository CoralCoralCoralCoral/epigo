package model

import "github.com/google/uuid"

type Config struct {
	Id uuid.UUID `json:"id"`

	// Global Params
	TimeStep  int64 `json:"time_step"`
	NumAgents int64 `json:"num_agents"`

	// Agent Params
	ComplianceProbability        float64 `json:"compliance_probability"`
	SeeksTreatmentProbability    float64 `json:"seeks_treatment_probability"`
	MaskFiltrationEfficiencyMean float64 `json:"mask_filtration_efficiency_mean"`
	MaskFiltrationEfficiencySd   float64 `json:"mask_filtration_efficiency_sd"`
	PulmonaryVentilationRateMean float64 `json:"pulmonary_ventilation_rate_mean"`
	PulmonaryVentilationRateSd   float64 `json:"pulmonary_ventilation_rate_sd"`

	// Pathogen Params
	IncubationPeriodMean         float64 `json:"incubation_period_mean"`
	IncubationPeriodSd           float64 `json:"incubation_period_sd"`
	RecoveryPeriodMean           float64 `json:"recovery_period_mean"`
	RecoveryPeriodSd             float64 `json:"recovery_period_sd"`
	ImmunityPeriodMean           float64 `json:"immunity_period_mean"`
	ImmunityPeriodSd             float64 `json:"immunity_period_sd"`
	PrehospitalizationPeriodMean float64 `json:"prehospitalization_period_mean"`
	PrehospitalizationPeriodSd   float64 `json:"prehospitalization_period_sd"`
	HospitalizationPeriodMean    float64 `json:"hospitalization_period_mean"`
	HospitalizationPeriodSd      float64 `json:"hospitalization_period_sd"`
	QuantaEmissionRateMean       float64 `json:"quanta_emission_rate_mean"`
	QuantaEmissionRateSd         float64 `json:"quanta_emission_rate_sd"`
	HospitalizationProbability   float64 `json:"hospitalization_probability"`
	DeathProbability             float64 `json:"death_probability"` // conditional on hospitalized
	AsymptomaticProbability      float64 `json:"asymptomatic_probability"`

	// Household Params
	HouseholdCapacityMean      float64 `json:"household_capacity_mean"`
	HouseholdCapacitySd        float64 `json:"household_capacity_sd"`
	HouseholdAirChangeRateMean float64 `json:"household_air_change_rate_mean"`
	HouseholdAirChangeRateSd   float64 `json:"household_air_change_rate_sd"`
	HouseholdVolumeMean        float64 `json:"household_volume_mean"`
	HouseholdVolumeSd          float64 `json:"household_volume_sd"`

	// Office Params
	OfficeCapacityMean      float64 `json:"office_capacity_mean"`
	OfficeCapacitySd        float64 `json:"office_capacity_sd"`
	OfficeAirChangeRateMean float64 `json:"office_air_change_rate_mean"`
	OfficeAirChangeRateSd   float64 `json:"office_air_change_rate_sd"`
	OfficeVolumeMean        float64 `json:"office_volume_mean"`
	OfficeVolumeSd          float64 `json:"office_volume_sd"`

	// Social Space Params
	SocialSpaceCapacityMean      float64 `json:"social_space_capacity_mean"`
	SocialSpaceCapacitySd        float64 `json:"social_space_capacity_sd"`
	SocialSpaceAirChangeRateMean float64 `json:"social_space_air_change_rate_mean"`
	SocialSpaceAirChangeRateSd   float64 `json:"social_space_air_change_rate_sd"`
	SocialSpaceVolumeMean        float64 `json:"social_space_volume_mean"`
	SocialSpaceVolumeSd          float64 `json:"social_space_volume_sd"`

	// Healthcare Space Params
	HealthcareSpaceCapacityMean      float64 `json:"healthcare_space_capacity_mean"`
	HealthcareSpaceCapacitySd        float64 `json:"healthcare_space_capacity_sd"`
	HealthcareSpaceAirChangeRateMean float64 `json:"healthcare_space_air_change_rate_mean"`
	HealthcareSpaceAirChangeRateSd   float64 `json:"healthcare_space_air_change_rate_sd"`
	HealthcareSpaceVolumeMean        float64 `json:"healthcare_space_volume_mean"`
	HealthcareSpaceVolumeSd          float64 `json:"healthcare_space_volume_sd"`
	TestCapacityMean                 float64 `json:"test_capacity_mean"`
	TestCapacitySd                   float64 `json:"test_capacity_sd"`
	TestSensitivity                  float64 `json:"test_sensitivity"`
	TestSpecificity                  float64 `json:"test_specificity"`
}
