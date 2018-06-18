# Transitioner, golang FSM

Supports Events, Transitions, Guards and Callbacks (on global FSM and Transition levels).

## Usage

```go
package main

import (
  tr "github.com/dizer/transitioner"
  "fmt"
)

type Job struct {
  State    string
}

func (job *Job) Notify() error {
  fmt.Println("Notifying...")
  return nil
}

func main() {
  job := Job{}

  fsm := tr.Init(
    tr.FSMDescription{
      Initial: "sleeping",
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
              Callbacks: tr.Callbacks{
                After: []tr.CallbackFunc{job.Notify},
              },
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

  fsm.Bind(&job, "State")

  fmt.Println(job.State) // sleeping

  fsm.Fire("run")
  fmt.Println(job.State) // running

  fsm.Fire("stop")
  fmt.Println(job.State) // running

  fsm.Fire("toggle")
  fmt.Println(job.State) // sleeping
}
```

## Transactional rollbacks

You can use Transitioner with your SQL Lib to automatically rollback transactions on state transition failures.
Please watch examples dir for more details.
