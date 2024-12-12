package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSerializeAndDeserializeApplyPolicyUpdateCommand(t *testing.T) {
	is_mask_mandate := false
	is_lockdown := false
	test_strategy := TestEveryone

	command := Command{
		Type: ApplyPolicyUpdate,
		Payload: ApplyPolicyUpdatePayload{
			JurisdictionId: "GLOBAL",
			IsMaskMandate:  &is_mask_mandate,
			IsLockdown:     &is_lockdown,
			TestStrategy:   &test_strategy,
		},
	}

	commandBytes, err := json.Marshal(command)
	if err != nil {
		t.Fatalf("Test failed due to the following Marshalling error: %s", err)
	}

	var deserializedCommand Command
	err = json.Unmarshal(commandBytes, &deserializedCommand)
	if err != nil {
		t.Fatalf("Test failed due to the following Unmarshalling error: %s", err)
	}

	if deserializedCommand.Type != ApplyPolicyUpdate {
		t.Fatalf("Test failed because the unmarshalled command is not of the expected type. Expected %s, got %s", ApplyPolicyUpdate, command.Type)
	}

	assert.Equal(t, command.Type, deserializedCommand.Type, "Expected the unmarshalled command to have the same type as the original")
}

func TestSerializeAndDeserializeCommandsWithNoPayload(t *testing.T) {
	command := Command{
		Type: Quit,
	}

	commandBytes, err := json.Marshal(command)
	if err != nil {
		t.Fatalf("Test failed due to the following Marshalling error: %s", err)
	}

	var deserializedCommand Command
	err = json.Unmarshal(commandBytes, &deserializedCommand)
	if err != nil {
		t.Fatalf("Test failed due to the following Unmarshalling error: %s", err)
	}

	assert.Equal(t, command.Type, deserializedCommand.Type, "Expected the unmarshalled command to have the same type as the original")
}

func TestDeserializeCommandWithNoPayloadFromJsonString(t *testing.T) {
	commandBytes := []byte(`
		{
			"type": "pause"
		}
	`)

	var command Command
	err := json.Unmarshal(commandBytes, &command)
	if err != nil {
		t.Fatalf("Test failed due to the following Unmarshalling error: %s", err)
	}

	assert.Equal(t, Pause, command.Type, "Expected the unmarshalled command to have the same type as the original")
}
