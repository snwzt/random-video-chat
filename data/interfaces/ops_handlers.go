package interfaces

type UserOperationsHandler interface {
	UserMatcher()
	UserRemove()
}

type ForwarderOperationsHandler interface {
	CreateMatch()
	DeleteMatch()
}
