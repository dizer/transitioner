package transitioner

type CallbackFunc func() error

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
	EarlyBefore []CallbackFunc
	Before      []CallbackFunc
	LateBefore  []CallbackFunc
	EarlyAfter  []CallbackFunc
	After       []CallbackFunc
	LateAfter   []CallbackFunc
}

func (transition *Transition) Apply(fsm *FSM) error {
	var err error
	complete := false
	originalState := fsm.CurrentState

	defer func() {
		if !complete {
			fsm.CurrentState = originalState
		}
	}()

	// beforeCallbacks := append(fsm.Description.Callbacks.Before, transition.Callbacks.Before...)
	// for _, cb := range beforeCallbacks {
	// 	err := cb()
	// 	if err != nil {
	// 		return wrapCallbackError(err, cb)
	// 	}
	// }

	beforeCallbacks := [][]CallbackFunc{
		fsm.Description.Callbacks.EarlyBefore,
		transition.Callbacks.EarlyBefore,
		fsm.Description.Callbacks.Before,
		transition.Callbacks.Before,
		fsm.Description.Callbacks.LateBefore,
		transition.Callbacks.LateBefore,
	}

	for _, pack := range beforeCallbacks {
		err = runCallbacks(pack)
		if err != nil {
			return err
		}
	}

	fsm.CurrentState = transition.To

	afterCallbacks := [][]CallbackFunc{
		transition.Callbacks.EarlyAfter,
		fsm.Description.Callbacks.EarlyAfter,
		transition.Callbacks.After,
		fsm.Description.Callbacks.After,
		transition.Callbacks.LateAfter,
		fsm.Description.Callbacks.LateAfter,
	}

	for _, pack := range afterCallbacks {
		err = runCallbacks(pack)
		if err != nil {
			return err
		}
	}

	complete = true

	return nil
}

func (transition *Transition) IsValid(fsm *FSM) bool {
	allowedFrom := contains(transition.From, fsm.CurrentState)
	if !allowedFrom {
		return false
	}

	for _, guard := range transition.Guards {
		if !guard.If() {
			return false
		}
	}

	return true
}

func runCallbacks(cbs []CallbackFunc) error {
	for _, cb := range cbs {
		err := cb()
		if err != nil {
			return wrapCallbackError(err, cb)
		}
	}
	return nil
}
