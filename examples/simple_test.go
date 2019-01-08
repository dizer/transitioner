package main

import (
	"fmt"
	tr "github.com/dizer/transitioner"
)

type Job struct {
	State string
}

func (job *Job) CanStop() bool {
	return false
}

func ExampleStates(){
	job := Job{}

	fsm := tr.GetFSM(
		tr.FSMDescription{
			InitialState: "sleeping",
			Events: []tr.Event{

				{
					Name: "toggle",
					Transitions: []tr.Transition{
						// Will be used first suitable transition
						{From: []string{"sleeping"}, To: "running"},
						{From: []string{"running"}, To: "sleeping"},
					},
				},

				{
					Name: "run",
					Transitions: []tr.Transition{
						{
							From: []string{"sleeping"},
							To:   "running",
						},
					},
				},

				{
					Name: "stop",
					Transitions: []tr.Transition{
						{
							From: []string{"running"},
							To:   "sleeping",
							// event "stop" will never transit from running to sleeping
							Guards: []tr.Guard{
								{If: job.CanStop},
							},
						},
					},
				},
			},
		},
	)

	fsm.Bind(&job.State)
	fmt.Println(job.State)

	fsm.Fire("run")
	fmt.Println(job.State)

	fsm.Fire("stop")
	fmt.Println(job.State)

	fsm.Fire("toggle")
	fmt.Println(job.State)

	// Output:
	// sleeping
	// running
	// running
	// sleeping
}
