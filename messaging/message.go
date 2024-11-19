package messaging

type Message struct {
	Payload interface{}
}

// func (msg Message) MarshalJSON() ([]byte, error) {
// 	payload := []byte{}

// 	switch msg.Payload.(type) {
// 	case string:
// 		payload = []byte(payload)
// 	}

// 	// Create a map to customize the JSON structure
// 	customData := map[string]interface{}{
// 		"api_id":        msg.ApiId,
// 		"sim_server_id": msg.SimServerId,
// 		"sim_id":        msg.SimId,
// 		"payload":       payload,
// 	}

// 	// Serialize the customData map to JSON
// 	return json.Marshal(customData)
// }
