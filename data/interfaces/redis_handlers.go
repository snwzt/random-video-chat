package interfaces

type UserQueueHandler interface {
	Matcher()
	UserRemove()
}

type ForwarderHandler interface {
	CreateMatch()
	DeleteMatch()
}
