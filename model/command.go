package model

const Quit CommandType = "quit"
const Pause CommandType = "pause"
const Resume CommandType = "resume"
const ApplyJurisdictionPolicy CommandType = "apply_jurisdiction_policy"

type Command struct {
	Type    CommandType `json:"type"`
	Payload interface{} `json:"payload"`
}

type CommandType string

type ApplyJurisdictionPolicyPayload struct {
	JurisdictionId string `json:"jurisdiction_id"`
	Policy         Policy `json:"policy"`
}
