package model

const TestEveryone TestStrategy = "everyone"
const TestSymptomatic TestStrategy = "symptomatic"
const TestNone TestStrategy = "none"

type Policy struct {
	IsMaskMandate bool         `json:"is_mask_mandate"`
	IsLockDown    bool         `json:"is_lockdown"`
	TestStrategy  TestStrategy `json:"test_strategy"`
}

type TestStrategy string
