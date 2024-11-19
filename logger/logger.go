package logger

import "github.com/CoralCoralCoralCoral/simulation-engine/protos/protos"

type Logger struct {
	events   chan *protos.Event
	channels []chan *protos.Event
}

func NewLogger() Logger {
	return Logger{
		events:   make(chan *protos.Event),
		channels: make([]chan *protos.Event, 0),
	}
}

func (logger *Logger) Log(event *protos.Event) {
	logger.events <- event
}

func (logger *Logger) Subscribe(subscriber func(event *protos.Event)) {
	channel := make(chan *protos.Event)

	logger.channels = append(logger.channels, channel)

	go func() {
		for {
			event := <-channel
			subscriber(event)
		}
	}()
}

func (logger *Logger) Broadcast() {
	for {
		event := <-logger.events

		for _, channel := range logger.channels {
			channel <- event
		}
	}
}
