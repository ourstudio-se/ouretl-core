package core

type defaultDataMessage struct {
	id     string
	data   []byte
	origin string
}

func (dm *defaultDataMessage) ID() string {
	return dm.id
}

func (dm *defaultDataMessage) Data() []byte {
	return dm.data
}

func (dm *defaultDataMessage) Origin() string {
	return dm.origin
}

func (dm *defaultDataMessage) withData(data []byte) *defaultDataMessage {
	dm.data = data
	return dm
}
