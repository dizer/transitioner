package transitioner

import (
	"fmt"
	"reflect"
)

type CallbackFunc func() error
type CallbackAroundPassFunc func(error) error
type CallbackAroundFunc func(CallbackAroundPassFunc) CallbackAroundPassFunc

type FSM struct {
	Object       interface{}
	Description  FSMDescription
	Current      string
	CurrentEvent *string
}

type FSMDescription struct {
	Initial   string
	Events    []Event
	Callbacks Callbacks
}

type Event struct {
	Name        string
	Transitions []Transition
}

type Transition struct {
	From      []string
	To        string
	Guards    []Guard
	Callbacks Callbacks
}

type Guard struct {
	If func() bool
}

type Callbacks struct {
	Before []CallbackFunc
	After  []CallbackFunc
	Around []CallbackAroundFunc
}

func Init(desc FSMDescription) *FSM {
	fsm := FSM{
		Description: desc,
	}
	fsm.Load()
	return &fsm
}

func (fsm *FSM) Bind(object interface{}, bindField string) {
	field := reflect.ValueOf(object).Elem().FieldByName(bindField)
	current := field.Interface().(string)

	if current != "" {
		fsm.Current = current
	}

	fsm.Object = object

	cb := func() error {
		field.SetString(fsm.Current)
		return nil
	}
	cb()
	fsm.Description.Callbacks.After = append(fsm.Description.Callbacks.After, cb)
}

func (fsm *FSM) Load() {
	fsm.Current = fsm.Description.Initial
}

func (fsm *FSM) State() string {
	return fsm.Current
}

func (fsm *FSM) Fire(eventName string) (err error) {
	if fsm == nil {
		return fmt.Errorf("please init FSM first")
	}

	fsm.CurrentEvent = &eventName

	defer func() {
		fsm.CurrentEvent = nil
	}()

	eventDesc, err := fsm.getEventDesc(eventName)

	if err != nil {
		return err
	}

	for _, transition := range eventDesc.Transitions {
		if transition.IsValid(fsm) {
			return transition.Apply(fsm)
		}
	}

	return fmt.Errorf("no available transitions found for event '%s' from  state '%s'", eventName, fsm.Current)
}

func (transition *Transition) Apply(fsm *FSM) error {
	fn := func(_ error) error {
		beforeCallbacks := append(fsm.Description.Callbacks.Before, transition.Callbacks.Before...)
		for _, fn := range beforeCallbacks {
			err := fn()
			if err != nil {
				return err
			}
		}

		fsm.Current = transition.To

		afterCallbacks := append(fsm.Description.Callbacks.After, transition.Callbacks.After...)
		for _, fn := range afterCallbacks {
			err := fn()
			if err != nil {
				return err
			}
		}

		return nil
	}

	aroundCbs := append(fsm.Description.Callbacks.Around, transition.Callbacks.Around...)

	cbStack := make([]func(error) error, len(aroundCbs)+1)
	cbStack[0] = fn

	for i, piece := range aroundCbs {
		cbStack[i+1] = piece(cbStack[i])
	}

	err := cbStack[len(aroundCbs)](nil)
	return err
}

func (transition *Transition) IsValid(fsm *FSM) bool {
	inFrom := contains(transition.From, fsm.Current)
	if !inFrom {
		return false
	}

	for _, guard := range transition.Guards {
		if !guard.If() {
			return false
		}
	}

	return true
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

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
