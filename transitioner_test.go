package transitioner

import (
	"fmt"
	"testing"
)

type testObj struct {
	Field        string
	AnotherField string
	FSM          *FSM
}

func (obj *testObj) TrueGuard(*FSM) bool {
	return true
}

func TestInitialState(t *testing.T) {
	fsm := GetFSM(
		FSMDescription{
			InitialState: "initialized",
		},
	)

	if fsm.State() != "initialized" {
		t.Error("expected state to be 'initialized'")
	}

}

func TestFireEventNotExists(t *testing.T) {
	fsm := GetFSM(
		FSMDescription{
			InitialState: "initialized",
		},
	)
	err := fsm.Fire("some-event")
	if err == nil {
		t.Error("fire unknown event returns an error")
	}
}

func TestFireEvent(t *testing.T) {
	fsm := GetFSM(
		FSMDescription{
			InitialState: "initialized",
			Events: []Event{
				{
					Name: "start",
					Transitions: []Transition{
						{
							From: []string{"initialized"},
							To:   "started",
							Guards: []Guard{
								{If: func() bool { return false }},
							},
						},
						{
							From: []string{"initialized"},
							To:   "failed",
						},
					},
				},
			},
		},
	)

	err := fsm.Fire("start")

	if err != nil {
		t.Error(err)
	}

	if fsm.State() != "failed" {
		t.Errorf("expected state to be '%s'", "failed")
	}
}

func TestBind(t *testing.T) {
	var err error
	obj := testObj{Field: "unchanged"}

	fsm := GetFSM(
		FSMDescription{
			InitialState: "initialized",
			Events: []Event{
				{
					Name: "start",
					Transitions: []Transition{
						{From: []string{"unchanged", "initialized"}, To: "started"},
					},
				},
			},
		},
	)

	err = fsm.Bind(&obj.Field)
	if err != nil {
		t.Error(err)
	}


	if obj.Field != "unchanged" {
		t.Error("Bind() should not change Field from 'unchanged'")
	}

	err = fsm.Fire("start")

	if err != nil {
		t.Error(err)
	}

	if obj.Field != "started" {
		t.Errorf("Bound field should be updated to '%s', but got '%s'", "started", obj.Field)
	}
}

func TestGuardsAllow(t *testing.T) {
	var err error
	obj := testObj{Field: "", AnotherField: "pass guard"}

	fsm := GetFSM(
		FSMDescription{
			InitialState: "initialized",
			Events: []Event{
				{
					Name: "start",
					Transitions: []Transition{
						{
							From: []string{"initialized"},
							To:   "started",
							Guards: []Guard{
								{If: func() bool {
									return obj.AnotherField == "pass guard"
								}},
							},
						},
					},
				},
			},
		},
	)

	err = fsm.Bind(&obj.Field)
	if err != nil {
		t.Error(err)
	}

	err = fsm.Fire("start")
	if err != nil {
		t.Error(err)
	}

	if fsm.State() != "started" {
		t.Error("Guards should allow transition")
	}

}

func TestGuardsDisallow(t *testing.T) {
	var err error
	obj := testObj{Field: "", AnotherField: "pass guard"}

	fsm := GetFSM(
		FSMDescription{
			InitialState: "initialized",
			Events: []Event{
				{
					Name: "start",
					Transitions: []Transition{
						{
							From: []string{"initialized"},
							To:   "started",
							Guards: []Guard{
								{If: func() bool {
									return obj.AnotherField != "pass guard"
								}},
							},
						},
					},
				},
			},
		},
	)

	err = fsm.Bind(&obj.Field)
	if err != nil {
		t.Error(err)
	}

	err = fsm.Fire("start")
	if err == nil {
		t.Error(err)
	}

	if fsm.State() == "started" {
		t.Error("Guards should not allow transition")
	}
}

func TestCallbacks(t *testing.T) {
	var err error
	obj := testObj{Field: ""}

	fsm := GetFSM(
		FSMDescription{
			InitialState: "initialized",
			Events: []Event{
				{
					Name: "start",
					Transitions: []Transition{
						{
							From: []string{"initialized"},
							To:   "failed",
							Callbacks: Callbacks{
								After: []CallbackFunc{
									func() error {
										obj.AnotherField = "failure data"
										return nil
									},
								},
							},
						},
					},
				},
			},
		},
	)

	err = fsm.Bind(&obj.Field)
	if err != nil {
		t.Error(err)
	}

	err = fsm.Fire("start")
	if err != nil {
		t.Error(err)
	}

	if obj.AnotherField != "failure data" {
		t.Error("Callbacks should be executes")
	}
}

func TestAfterCallbacks(t *testing.T) {
	var err error
	obj := new(testObj)

	fsm := GetFSM(
		FSMDescription{
			InitialState: "initialized",
			Events: []Event{
				{
					Name: "run",
					Transitions: []Transition{
						{
							From: []string{"initialized"},
							To:   "succeed",
							Callbacks: Callbacks{
								After: []CallbackFunc{
									func() error { return fmt.Errorf("example error")},
								},
							},
						},
					},
				},
			},
		},
	)

	err = fsm.Bind(&obj.Field)
	if err != nil {
		t.Error(err)
	}

	err = fsm.Fire("run")
	if err == nil {
		t.Error(err)
	}

	if fsm.State() != "initialized" {
		t.Error("Callback error prevent transition")
	}
}
