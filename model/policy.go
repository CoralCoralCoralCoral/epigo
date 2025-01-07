package model

const TestEveryone TestStrategy = "everyone"
const TestSymptomatic TestStrategy = "symptomatic"
const TestNone TestStrategy = "none"

type TestStrategy string

type Policy struct {
	IsMaskMandate          bool         `json:"is_mask_mandate"`
	IsSelfIsolationMandate bool         `json:"is_self_isolation_mandate"`
	IsSelfReportingMandate bool         `json:"is_self_reporting_mandate"`
	IsLockdown             bool         `json:"is_lockdown"`
	TestStrategy           TestStrategy `json:"test_strategy"`
	TestCapacityMultiplier float64      `json:"test_capacity_multiplier"`
	ComplianceProbability  float64      `json:"compliance_probability"`
}
