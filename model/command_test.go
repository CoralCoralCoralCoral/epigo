package model

import (
	"encoding/json"
	"testing"
)

func TestSerializeAndDeserializeCommand(t *testing.T) {
	command := Command{
		Type: ApplyJurisdictionPolicy,
		Payload: ApplyJurisdictionPolicyPayload{
			JurisdictionId: "GLOBAL",
			Policy: Policy{
				IsMaskMandate: false,
				IsLockDown:    false,
				TestStrategy:  TestEveryone,
			},
		},
	}

	commandBytes, err := json.Marshal(command)
	if err != nil {
		t.Fatalf("Test failed due to the following Marshalling error: %s", err)
	}

	var deserializedCommand Command
	json.Unmarshal(commandBytes, &deserializedCommand)

	if deserializedCommand.Type != ApplyJurisdictionPolicy {
		t.Fatalf("Test failed because the unmarshalled command is not of the expected type. Expected %s, got %s", ApplyJurisdictionPolicy, command.Type)
	}

	if _, ok := deserializedCommand.Payload.(*ApplyJurisdictionPolicyPayload); !ok {
		t.Fatalf("Test failed because the deserialized command does not contain the expected payload")
	}
}
