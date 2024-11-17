package logger

type Event struct {
	Type    EventType
	Payload interface{}
}

type EventType string
