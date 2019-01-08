package transitioner

import "fmt"

type FSM struct {
	Description  FSMDescription
	CurrentState string
}

func (fsm *FSM) Bind(bindField *string) error {
	if bindField != nil && *bindField != "" {
		fsm.CurrentState = *bindField
	}

	cb := func() error {
		*bindField = fsm.CurrentState
		return nil
	}

	err := cb()
	if err != nil {
		return err
	}

	fsm.Description.Callbacks.EarlyAfter = append([]CallbackFunc{cb}, fsm.Description.Callbacks.EarlyAfter...)
	return nil
}

func (fsm *FSM) Initialize() {
	fsm.CurrentState = fsm.Description.InitialState
}

func (fsm *FSM) State() string {
	return fsm.CurrentState
}

func (fsm *FSM) Fire(eventName string) (err error) {
	if fsm == nil {
		return fmt.Errorf("please init FSM first")
	}

	eventDesc, err := fsm.getEventDesc(eventName)

	if err != nil {
		return err
	}

	for _, transition := range eventDesc.Transitions {
		if transition.IsValid(fsm) {
			return transition.Apply(fsm)
		}
	}

	return fmt.Errorf("no available transitions found for event '%s' from  state '%s'", eventName, fsm.CurrentState)
}

func (fsm *FSM) getEventDesc(eventName string) (eventDesc *Event, err error) {
	for _, e := range fsm.Description.Events {
		if e.Name == eventName {
			eventDesc = &e
			break
		}
	}

	if eventDesc == nil {
		return nil, fmt.Errorf("event %s not described", eventName)
	}

	return eventDesc, nil
}
