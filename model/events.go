package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/umran/epigo/logger"
)

const EpochEnd logger.EventType = "epoch_end"
const CommandProcessed logger.EventType = "command_processed"
const AgentStateUpdate logger.EventType = "agent_state_update"
const AgentLocationUpdate logger.EventType = "agent_location_update"
const SpaceOccupancyUpdate logger.EventType = "space_occupancy_update"

type EpochEndPayload struct {
	Epoch    int64
	TimeStep int64
	Time     time.Time
}

type CommandProcessedPayload struct {
	Epoch   int64
	Command Command
}

type AgentStateUpdatePayload struct {
	Epoch int64
	Id    uuid.UUID
	State AgentState
}

type AgentLocationUpdatePayload struct {
	Epoch      int64
	Id         uuid.UUID
	LocationId uuid.UUID
}

type SpaceOccupancyUpdatePayload struct {
	Epoch     int64
	Id        uuid.UUID
	Occupants []struct {
		Id    uuid.UUID
		State AgentState
	}
}
