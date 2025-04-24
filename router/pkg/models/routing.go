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

func (m RoutingMessage) Dump() []byte {
	panic("not impl")
}

type RoutingReply struct {
	DeviceMac string // net.HardwareAddr
	Ack       bool
}

func (m RoutingReply) Dump() []byte {
	panic("not impl")
}

func (m RoutingReply) FromMqtt(payload []byte) (RoutingReply, error) {
	panic("not impl")
}
