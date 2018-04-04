package transitioner

import "github.com/jinzhu/gorm"

func (fsm *FSM) IntegrateGorm(db *gorm.DB) {
	transactionCb := func(fsm *FSM, fn func(*FSM, error) (*FSM, error)) func(*FSM, error) (*FSM, error) {
		return func(fsm *FSM, err error) (*FSM, error) {
			if err != nil {
				return fsm, err
			}
			tx := db.Begin()
			retFSM, retERR := fn(fsm, err)
			if retERR != nil {
				tx.Rollback()
			} else {
				tx.Commit()
			}
			return retFSM, retERR
		}
	}

	fsm.Desc.Callbacks.Around = append(fsm.Desc.Callbacks.Around, transactionCb)
}
