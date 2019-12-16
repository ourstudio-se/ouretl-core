package core

type DefaultDataMessage struct {
	id     string
	data   []byte
	origin string
}

func (dm *DefaultDataMessage) ID() string {
	return dm.id
}

func (dm *DefaultDataMessage) Data() []byte {
	return dm.data
}

func (dm *DefaultDataMessage) Origin() string {
	return dm.origin
}

func (dm *DefaultDataMessage) withData(data []byte) *DefaultDataMessage {
	dm.data = data
	return dm
}
