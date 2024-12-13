package model

import (
	"encoding/json"
)

const Quit CommandType = "quit"
const Pause CommandType = "pause"
const Resume CommandType = "resume"
const ApplyPolicyUpdate CommandType = "apply_policy_update"

type Command struct {
	Type    CommandType `json:"type"`
	Payload interface{} `json:"payload"`
}

type CommandType string

type ApplyPolicyUpdatePayload struct {
	JurisdictionId         string        `json:"jurisdiction_id"`
	IsMaskMandate          *bool         `json:"is_mask_mandate"`
	IsLockdown             *bool         `json:"is_lockdown"`
	TestStrategy           *TestStrategy `json:"test_strategy"`
	TestCapacityMultiplier *float64      `json:"test_capacity_multiplier"`
}

// UnmarshalJSON implements the custom unmarshalling logic for Command.
func (c *Command) UnmarshalJSON(data []byte) error {
	// Define an intermediate structure to capture the "type" and raw "payload".
	var intermediate struct {
		Type    CommandType      `json:"type"`
		Payload *json.RawMessage `json:"payload"`
	}

	// Unmarshal into the intermediate structure.
	if err := json.Unmarshal(data, &intermediate); err != nil {
		return err
	}

	c.Type = intermediate.Type

	// Determine the actual type of the payload based on the "type" field.
	if intermediate.Payload == nil {
		return nil
	}

	var payload interface{}
	switch intermediate.Type {
	case ApplyPolicyUpdate:
		payload = &ApplyPolicyUpdatePayload{}
	default:
		payload = &map[string]interface{}{}
	}

	if err := json.Unmarshal(*intermediate.Payload, payload); err != nil {
		return err
	}

	c.Payload = payload

	return nil
}
