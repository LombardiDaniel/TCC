package models

import "encoding/json"

type RequestType int

const (
	ImageRequest RequestType = iota
	BlinkRequest
)

type RoutingMessage struct {
	DeviceMac string            `json:"deviceMac"` // net.HardwareAddr
	Type      RequestType       `json:"type"`
	Fields    map[string]string `json:"fields"`
}

func (m RoutingMessage) Dump() []byte {
	jsonData, _ := json.Marshal(m)
	return jsonData
}

type RoutingReply struct {
	DeviceMac string `json:"deviceMac"` // net.HardwareAddr
	Ack       bool   `json:"ack"`
}

func (m RoutingReply) Dump() []byte {
	jsonData, _ := json.Marshal(m)
	return jsonData
}

func (m RoutingReply) FromMqtt(payload []byte) (RoutingReply, error) {
	var reply RoutingReply
	err := json.Unmarshal(payload, &reply)
	if err != nil {
		return RoutingReply{}, err
	}
	return reply, nil
}
