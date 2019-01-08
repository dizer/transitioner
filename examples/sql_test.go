package main

import (
	"database/sql"
	"fmt"
	tr "github.com/dizer/transitioner"
	_ "github.com/mattn/go-sqlite3"
)

type JobWithDB struct {
	ID    int
	State string
}

func prepare() (*sql.DB) {
	db, err := sql.Open("sqlite3", "file:example.sqlite?cache=shared&mode=memory")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("CREATE TABLE jobs (id INTEGER PRIMARY KEY, state TEXT NOT NULL)")
	if err != nil {
		panic(err)
	}

	return db
}

func ExampleSQL() {
	var (
		err   error
		state string
	)

	db := prepare()
	defer db.Close()

	job := JobWithDB{
		ID:    1,
		State: "sleeping",
	}
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	err = job.Save(tx)()
	if err != nil {
		panic(err)
	}
	err = tx.Commit()
	if err != nil {
		panic(err)
	}

	err = db.QueryRow("select state from jobs where id = ?", job.ID).Scan(&state)
	if err != nil {
		panic(err)
	}
	fmt.Println(state)

	err = job.FireTransactional("toggle", db)
	if err != nil {
		panic(err)
	}

	err = db.QueryRow("select state from jobs where id = ?", job.ID).Scan(&state)
	if err != nil {
		panic(err)
	}
	fmt.Println(state)

	// Output:
	// sleeping
	// running
}

func (job *JobWithDB) FireTransactional(eventName string, db *sql.DB) error {
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

	fsm, err := job.BindFSM(tx)

	err = fsm.Fire(eventName)
	if err != nil {
		tx.Rollback()
		return err
	} else {
		return tx.Commit()
	}
}

func (job *JobWithDB) Save(tx *sql.Tx) func() error {
	return func() error {
		// upsert requires sqlite 3.24.0 (mattn/go-sqlite3 v1.9.0+)
		_, err := tx.Exec("INSERT INTO jobs(id, state) VALUES(?, ?) ON CONFLICT(jobs.id) DO UPDATE SET state=? WHERE id = ?", job.ID, job.State, job.State, job.ID)
		return err
	}
}

func (job *JobWithDB) BindFSM(tx *sql.Tx) (*tr.FSM, error) {
	fsm := tr.GetFSM(
		tr.FSMDescription{
			InitialState: "sleeping",
			Events: []tr.Event{
				{
					Name: "toggle",
					Transitions: []tr.Transition{
						{
							From: []string{"sleeping"},
							To:   "running",
							Callbacks: tr.Callbacks{
								After: []tr.CallbackFunc{job.Save(tx)},
							},
						},
						{
							From: []string{"running"},
							To:   "sleeping",
							Callbacks: tr.Callbacks{
								After: []tr.CallbackFunc{job.Save(tx)},
							},
						},
					},
				},
			},
		},
	)
	return fsm, fsm.Bind(&job.State)
}
