package messaging

const FeedbackUpdate UpdateType = "feedback"
const MetricsUpdate UpdateType = "metrics"

type Update struct {
	Type    UpdateType  `json:"type"`
	Payload interface{} `json:"payload"`
}

type UpdateType string
