package transitioner

import (
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
	fsm := Init(
		FSMDesc{
			Initial: "initialized",
		},
	)

	if fsm.State() != "initialized" {
		t.Error("expected state to be 'initialized'")
	}

}

func TestFireEventNotExists(t *testing.T) {
	fsm := Init(
		FSMDesc{
			Initial: "initialized",
		},
	)
	err := fsm.Fire("some-event")
	if err == nil {
		t.Error("undescribed event should give an error")
	}
}

func TestFireEvent(t *testing.T) {
	fsm := Init(
		FSMDesc{
			Initial: "initialized",
			Events: []EventDesc{
				{
					Name: "start",
					Transitions: []TransitionDesc{
						{
							From: []string{"initialized"},
							To:   "started",
							Guards: []GuardDesc{
								{If: func(fsm *FSM) bool { return false }},
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

	fsm.Fire("start")

	if fsm.State() != "failed" {
		t.Error("expected state to be 'failed'")
	}
}

func TestBind(t *testing.T) {
	obj := testObj{Field: "unchanged"}

	fsm := Init(
		FSMDesc{
			Initial: "initialized",
			Events: []EventDesc{
				{
					Name: "start",
					Transitions: []TransitionDesc{
						{From: []string{"initialized"}, To: "started"},
					},
				},
			},
		},
	)

	fsm.Bind(&obj, "Field")

	if obj.Field != "initialized" {
		t.Error("Bind() should set Field to 'initialized'")
	}

	fsm.Fire("start")

	if obj.Field != "started" {
		t.Error("Binded field should be updated to 'started'")
	}
}

func TestGuardsAllow(t *testing.T) {
	obj := testObj{Field: "unchanged", AnotherField: "pass guard"}

	fsm := Init(
		FSMDesc{
			Initial: "initialized",
			Events: []EventDesc{
				{
					Name: "start",
					Transitions: []TransitionDesc{
						{
							From: []string{"initialized"},
							To:   "started",
							Guards: []GuardDesc{
								{If: func(fsm *FSM) bool {
									obj := fsm.Object.(*testObj)
									return obj.AnotherField == "pass guard"
								}},
							},
						},
					},
				},
			},
		},

	)

	fsm.Bind(&obj, "Field")

	fsm.Fire("start")

	if fsm.Current != "started" {
		t.Error("Guards shoud allow transition")
	}

}

func TestGuardsDisallow(t *testing.T) {
	obj := testObj{Field: "unchanged", AnotherField: "pass guard"}

	fsm := Init(
		FSMDesc{
			Initial: "initialized",
			Events: []EventDesc{
				{
					Name: "start",
					Transitions: []TransitionDesc{
						{
							From: []string{"initialized"},
							To:   "started",
							Guards: []GuardDesc{
								{If: func(fsm *FSM) bool {
									obj := fsm.Object.(*testObj)
									return obj.AnotherField != "pass guard"
								}},
							},
						},
					},
				},
			},
		},

	)

	fsm.Bind(&obj, "Field")

	fsm.Fire("start")

	if fsm.Current == "started" {
		t.Error("Guards shoud not allow transition")
	}

}

func TestCallbacks(t *testing.T) {
	obj := testObj{Field: "unchanged"}

	fsm := Init(
		FSMDesc{
			Initial: "initialized",
			Events: []EventDesc{
				{
					Name: "start",
					Transitions: []TransitionDesc{
						{
							From: []string{"initialized"},
							To:   "failed",
							Callbacks: CallbacksDesc{
								After: []func(*FSM){
									func(fsm *FSM) {
										obj := fsm.Object.(*testObj)
										obj.AnotherField = "failure data"
									},
								},
							},
						},
					},
				},
			},
		},

	)

	fsm.Bind(&obj, "Field")

	fsm.Fire("start")

	if obj.AnotherField != "failure data" {
		t.Error("Callbacks shoud be executes")
	}

}
