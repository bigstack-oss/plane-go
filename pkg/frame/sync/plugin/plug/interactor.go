package plug

var InteractPluggers = make(map[string]InteractPlugger)

type InteractPlugger interface {
	ConfigSetter
	ConfigChecker
	Stopper
}

type InteractUser interface {
	SetStage(string, int, interface{})
	SetInteractor(string)
	StartInteractor()
	StopInteractor()
}
