package messaging

const EventNotification NotificationType = "event"
const MetricsNotification NotificationType = "metrics"

type Notification struct {
	Type    NotificationType `json:"type"`
	Payload interface{}      `json:"payload"`
}

type NotificationType string
