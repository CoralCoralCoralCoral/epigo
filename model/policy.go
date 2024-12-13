package model

const TestEveryone TestStrategy = "everyone"
const TestSymptomatic TestStrategy = "symptomatic"
const TestNone TestStrategy = "none"

type TestStrategy string

type Policy struct {
	is_mask_mandate          bool
	is_lockdown              bool
	test_strategy            TestStrategy
	test_capacity_multiplier float64
}
