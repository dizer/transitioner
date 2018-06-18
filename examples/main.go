package main

import (
	"database/sql"
	"fmt"

	tr "github.com/dizer/transitioner"
)

type Job struct {
	State string
}

func (job *Job) Notify() error {
	fmt.Println("Notifying...")
	return nil
}

func (job *Job) CanStop() bool {
	return false
}

func (job *Job) Save(tx *sql.Tx) func() error {
	return func() error {
		_, err := tx.Exec("")
		return err
	}
}

// Example of auto-rollback on state change failure
func (job *Job) FireTransactional(eventName string, db sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	fsm := tr.Init(
		tr.FSMDescription{
			Initial: "sleeping",
			Events: []tr.Event{

				{
					Name: "toggle",
					Transitions: []tr.Transition{
						{From: []string{"sleeping"}, To: "running", Callbacks: tr.Callbacks{After: []tr.CallbackFunc{job.Save(tx)}}},
						{From: []string{"running"}, To: "sleeping", Callbacks: tr.Callbacks{After: []tr.CallbackFunc{job.Save(tx)}}},
					},
				},
			},
		},
	)

	err = fsm.Fire(eventName)
	if err != nil {
		tx.Rollback()
		return err
	} else {
		return tx.Commit()
	}
}

func main() {}
