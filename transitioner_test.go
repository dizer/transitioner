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
		FSMDescription{
			Initial: "initialized",
		},
	)

	if fsm.State() != "initialized" {
		t.Error("expected state to be 'initialized'")
	}

}

func TestFireEventNotExists(t *testing.T) {
	fsm := Init(
		FSMDescription{
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
		FSMDescription{
			Initial: "initialized",
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

	fsm.Fire("start")

	if fsm.State() != "failed" {
		t.Error("expected state to be 'failed'")
	}
}

func TestBind(t *testing.T) {
	obj := testObj{Field: "unchanged"}

	fsm := Init(
		FSMDescription{
			Initial: "initialized",
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

	fsm.Bind(&obj, "Field")

	if obj.Field != "unchanged" {
		t.Error("Bind() should not change Field from 'unchanged'")
	}

	fsm.Fire("start")

	if obj.Field != "started" {
		t.Error("Binded field should be updated to 'started'")
	}
}

func TestGuardsAllow(t *testing.T) {
	obj := testObj{Field: "", AnotherField: "pass guard"}

	fsm := Init(
		FSMDescription{
			Initial: "initialized",
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

	fsm.Bind(&obj, "Field")

	fsm.Fire("start")

	if fsm.Current != "started" {
		t.Error("Guards shoud allow transition")
	}

}

func TestGuardsDisallow(t *testing.T) {
	obj := testObj{Field: "", AnotherField: "pass guard"}

	fsm := Init(
		FSMDescription{
			Initial: "initialized",
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

	fsm.Bind(&obj, "Field")

	fsm.Fire("start")

	if fsm.Current == "started" {
		t.Error("Guards shoud not allow transition")
	}

}

func TestCallbacks(t *testing.T) {
	obj := testObj{Field: ""}

	fsm := Init(
		FSMDescription{
			Initial: "initialized",
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

	fsm.Bind(&obj, "Field")

	fsm.Fire("start")

	if obj.AnotherField != "failure data" {
		t.Error("Callbacks shoud be executes")
	}
}

func TestCallbacksChain(t *testing.T) {
	result := ""

	fsm := Init(
		FSMDescription{
			Initial: "initialized",
			Events: []Event{
				{
					Name: "start",
					Transitions: []Transition{
						{
							From: []string{"initialized"},
							To:   "running",
							Callbacks: Callbacks{
								Around: []CallbackAroundFunc{
									func(fn CallbackAroundPassFunc) CallbackAroundPassFunc {
										return func(err error) error {
											if err != nil {
												return err
											}
											result = result + "CB1 > "
											retERR := fn(err)
											result = result + " < CB1"
											return retERR
										}
									},
									func(fn CallbackAroundPassFunc) CallbackAroundPassFunc {
										return func(err error) error {
											if err != nil {
												return err
											}
											result = result + "CB2 > "
											retERR := fn(err)
											result = result + " < CB2"
											return retERR
										}
									},
								},
							},
						},
					},
				},
			},
		},
	)

	fsm.Fire("start")

	if fsm.Current != "running" {
		t.Error("Callbacks shoud stop transition")
	}

	if result != "CB2 > CB1 >  < CB1 < CB2" {
		t.Error("around callbacks shoud be stacked")
	}
}
