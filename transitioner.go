package transitioner

type FSMDescription struct {
	InitialState string
	Events       []Event
	Callbacks    Callbacks
}

type Event struct {
	Name        string
	Transitions []Transition
}

func GetFSM(desc FSMDescription) *FSM {
	fsm := FSM{
		Description: desc,
	}
	fsm.Initialize()
	return &fsm
}



