package model

const Quit CommandType = "quit"
const Pause CommandType = "pause"
const Resume CommandType = "resume"

type Command struct {
	Type    CommandType `json:"type"`
	Payload interface{} `json:"payload"`
}

type CommandType string
