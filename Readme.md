# Transitioner, golang FSM

Supports Events, Transitions, Guards and Callbacks (on global FSM and Transition levels).

## Usage

```go
package main

import (
    "github.com/dizer/transitioner"
    "fmt"
)

type Job struct {
    State    string
    FSM      *transitioner.FSM
}

func (job *Job) Notify() {
    fmt.Println("Notifying...")
}

func main() {
    job := Job{}

    fsm := transitioner.Init(
        transitioner.FSMDesc{
            Initial: "sleeping",
            Events: []transitioner.EventDesc{

                {
                    Name: "toggle",
                    Transitions: []transitioner.TransitionDesc{
                        // Will be used first suitable transition
                        {From: []string{"sleeping"}, To: "running"},
                        {From: []string{"running"}, To: "sleeping"},
                    },
                },

                {
                    Name: "run",
                    Transitions: []transitioner.TransitionDesc{
                        {
                            From: []string{"sleeping"},
                            To:   "running",
                            Callbacks: transitioner.CallbacksDesc{
                                After: []func(*transitioner.FSM) error {
                                    func(fsm *transitioner.FSM) error {
                                        job := fsm.Object.(*Job)
                                        job.Notify()
                                        return nil
                                    },
                                },
                            },
                        },
                    },
                },

                {
                    Name: "stop",
                    Transitions: []transitioner.TransitionDesc{
                        {
                            From: []string{"running"},
                            To:   "sleeping",
                            // event "stop" will never transit from running to sleeping
                            Guards: []transitioner.GuardDesc{
                                {If: func(fsm *transitioner.FSM) bool {
                                    return false
                                }},
                            },
                        },
                    },
                },
            },
        },
    )

    fsm.Bind(&job, "State")
    job.FSM = fsm

    fmt.Println(job.FSM.Current) // sleeping

    job.FSM.Fire("run")
    fmt.Println(job.FSM.Current) // running

    job.FSM.Fire("stop")
    fmt.Println(job.FSM.Current) // running

    job.FSM.Fire("toggle")
    fmt.Println(job.FSM.Current) // sleeping
}

```
