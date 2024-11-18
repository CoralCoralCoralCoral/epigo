package model

import "github.com/google/uuid"

type Config struct {
	Id        uuid.UUID `json:"id"`
	TimeStep  int64     `json:"time_step"`
	NumAgents int64     `json:"num_agents"`
	Pathogen  Pathogen  `json:"pathogen"`
}
