package models

type RequestType int

const (
	ImageRequest RequestType = iota
	BlinkRequest
)

type RoutingMessage struct {
	DeviceMac string // net.HardwareAddr

	Type   RequestType
	Fields map[string]string
}

type RoutingReply struct {
	DeviceMac string // net.HardwareAddr
	Ack       bool
}
