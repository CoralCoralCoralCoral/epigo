package logger

type Logger struct {
	events   chan Event
	channels []chan *Event
}

func NewLogger() Logger {
	return Logger{
		events:   make(chan Event),
		channels: make([]chan *Event, 0),
	}
}

func (logger *Logger) Log(event Event) {
	logger.events <- event
}

func (logger *Logger) Subscribe(subscriber func(event *Event)) {
	channel := make(chan *Event)

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
			channel <- &event
		}
	}
}
