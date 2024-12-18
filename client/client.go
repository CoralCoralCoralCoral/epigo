package main

import (
	"encoding/json"
	"log"

	"github.com/CoralCoralCoralCoral/simulation-engine/model"

	"github.com/google/uuid"
)

func main() {
	config := model.Config{
		Id: uuid.New(),

		// Global Params
		TimeStep:  15 * 60 * 1000,
		NumAgents: 150000,

		// Agent Params
		ComplianceProbability:        0.65,
		SeeksTreatmentProbability:    0.4,
		MaskFiltrationEfficiencyMean: 0.75,
		MaskFiltrationEfficiencySd:   0.2,
		PulmonaryVentilationRateMean: 0.36,
		PulmonaryVentilationRateSd:   0.01,

		// Pathogen Config
		IncubationPeriodMean:         3 * 24 * 60 * 60 * 1000,
		IncubationPeriodSd:           8 * 60 * 60 * 1000,
		RecoveryPeriodMean:           7 * 24 * 60 * 60 * 1000,
		RecoveryPeriodSd:             8 * 60 * 60 * 1000,
		ImmunityPeriodMean:           330 * 24 * 60 * 60 * 1000,
		ImmunityPeriodSd:             90 * 24 * 60 * 60 * 1000,
		PrehospitalizationPeriodMean: 3 * 24 * 60 * 60 * 1000,
		PrehospitalizationPeriodSd:   8 * 60 * 60 * 1000,
		HospitalizationPeriodMean:    7 * 24 * 60 * 60 * 1000,
		HospitalizationPeriodSd:      3 * 24 * 60 * 60 * 1000,
		QuantaEmissionRateMean:       500,
		QuantaEmissionRateSd:         150,
		HospitalizationProbability:   0.15,
		DeathProbability:             0.75,
		AsymptomaticProbability:      0.10,

		// Household Params
		HouseholdCapacityMean:      4,
		HouseholdCapacitySd:        2,
		HouseholdAirChangeRateMean: 7,
		HouseholdAirChangeRateSd:   1,
		HouseholdVolumeMean:        17,
		HouseholdVolumeSd:          2,

		// Office Params
		OfficeCapacityMean:      10,
		OfficeCapacitySd:        2,
		OfficeAirChangeRateMean: 20,
		OfficeAirChangeRateSd:   5,
		OfficeVolumeMean:        60,
		OfficeVolumeSd:          20,

		// Social Space Params
		SocialSpaceCapacityMean:      10,
		SocialSpaceCapacitySd:        2,
		SocialSpaceAirChangeRateMean: 20,
		SocialSpaceAirChangeRateSd:   5,
		SocialSpaceVolumeMean:        60,
		SocialSpaceVolumeSd:          10,

		// Healthcare Space Params
		HealthcareSpaceCapacityMean:      173,
		HealthcareSpaceCapacitySd:        25,
		HealthcareSpaceAirChangeRateMean: 20,
		HealthcareSpaceAirChangeRateSd:   5,
		HealthcareSpaceVolumeMean:        120,
		HealthcareSpaceVolumeSd:          30,
		TestCapacityMean:                 300,
		TestCapacitySd:                   150,
		TestSensitivity:                  0.7,
		TestSpecificity:                  0.999,
	}

	body, err := json.Marshal(config)
	if err != nil {
		log.Fatal("failed to serialize config")
	}

	println(string(body))
}
