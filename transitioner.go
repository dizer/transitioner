package transitioner

import (
	"fmt"
	"reflect"
)

type FSM struct {
	Object  interface{}
	Desc    FSMDesc
	Current string
	Gorm    GormIntegration
}

type FSMDesc struct {
	Initial   string
	Events    []EventDesc
	Callbacks CallbacksDesc
}

type EventDesc struct {
	Name        string
	Transitions []TransitionDesc
}

type TransitionDesc struct {
	From      []string
	To        string
	Guards    []GuardDesc
	Callbacks CallbacksDesc
}

type GuardDesc struct {
	If func(*FSM) bool
}

type CallbacksDesc struct {
	Before []func(*FSM) error
	After  []func(*FSM) error
	Around []func(*FSM, func(*FSM, error) (*FSM, error)) func(*FSM, error) (*FSM, error)
}

type OptionalParams struct {
	BindTo string
}

type BindParams struct {
	BindTo string
}

func (fsm *FSM) Bind(object interface{}, bindField string) {
	fsm.Object = object
	optional := OptionalParams{
		BindTo: bindField,
	}
	fsm.ApplyOptionals([]OptionalParams{optional})
}

func Init(desc FSMDesc, optionals ...OptionalParams) *FSM {
	fsm := FSM{
		Desc: desc,
	}
	fsm.Load()
	fsm.ApplyOptionals(optionals)
	return &fsm
}

func (fsm *FSM) ApplyOptionals(optionals []OptionalParams) {
	for _, optional := range optionals {
		if optional.BindTo != "" {
			cb := func(fsm *FSM) error {
				field := reflect.ValueOf(fsm.Object).Elem().FieldByName(optional.BindTo)
				field.SetString(fsm.Current)
				return nil
			}
			cb(fsm)
			fsm.Desc.Callbacks.After = append(fsm.Desc.Callbacks.After, cb)
		}
	}
}

func (fsm *FSM) Load() {
	fsm.Current = fsm.Desc.Initial
}

func (fsm *FSM) State() string {
	return fsm.Current
}

func (fsm *FSM) Fire(eventName string) (err error) {
	eventDesc, err := fsm.getEventDesc(eventName)

	if err != nil {
		return err
	}

	for _, transition := range eventDesc.Transitions {
		if transition.IsValid(fsm) {
			transition.Apply(fsm)
			break
		}
	}

	return nil
}

func (transition *TransitionDesc) Apply(fsm *FSM) {
	fn := func(fsm *FSM, _ error) (*FSM, error) {
		beforeCallbacks := append(fsm.Desc.Callbacks.Before, transition.Callbacks.Before...)
		for _, fn := range beforeCallbacks {
			err := fn(fsm)
			if err != nil {
				return fsm, err
			}
		}

		fsm.Current = transition.To

		afterCallbacks := append(fsm.Desc.Callbacks.After, transition.Callbacks.After...)
		for _, fn := range afterCallbacks {
			err := fn(fsm)
			if err != nil {
				return fsm, err
			}
		}

		return fsm, nil
	}

	aroundCbs := append(fsm.Desc.Callbacks.Around, transition.Callbacks.Around...)

	cbStack := make([]func(*FSM, error) (*FSM, error), len(aroundCbs)+1)
	cbStack[0] = fn

	for i, piece := range aroundCbs {
		cbStack[i+1] = piece(fsm, cbStack[i])
	}

	cbStack[len(aroundCbs)](fsm, nil)
}

func (transition *TransitionDesc) IsValid(fsm *FSM) bool {
	inFrom := contains(transition.From, fsm.Current)
	if !inFrom {
		return false
	}

	for _, guard := range transition.Guards {
		if !guard.If(fsm) {
			return false
		}
	}

	return true
}

func (fsm *FSM) getEventDesc(eventName string) (eventDesc *EventDesc, err error) {
	for _, e := range fsm.Desc.Events {
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
