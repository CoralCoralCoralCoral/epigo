package model

const TestEveryone TestStrategy = "everyone"
const TestSymptomatic TestStrategy = "symptomatic"
const TestNone TestStrategy = "none"

type Policy struct {
	IsMaskMandate          bool         `json:"is_mask_mandate"`
	IsLockDown             bool         `json:"is_lockdown"`
	TestStrategy           TestStrategy `json:"test_strategy"`
	TestCapacityMultiplier float64      `json:"test_capacity_multiplier"`
}

type TestStrategy string
