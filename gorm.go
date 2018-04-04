package transitioner

import (
	"github.com/jinzhu/gorm"
)

type GormIntegration struct {
	Tx *gorm.DB
}

func (fsm *FSM) IntegrateGorm(db *gorm.DB) {

	fsm.Gorm = GormIntegration{Tx: db}

	transactionCb := func(fsm *FSM, fn func(*FSM, error) (*FSM, error)) func(*FSM, error) (*FSM, error) {
		return func(fsm *FSM, err error) (*FSM, error) {
			if err != nil {
				return fsm, err
			}

			oldTx := fsm.Gorm.Tx
			tx := db.Begin()

			if tx.Error != nil {
				return fsm, tx.Error
			}

			fsm.Gorm.Tx = tx

			retFSM, retERR := fn(fsm, err)

			if retERR != nil {
				tx.Rollback()
			} else {
				tx.Commit()
			}

			fsm.Gorm.Tx = oldTx

			return retFSM, retERR
		}
	}

	fsm.Desc.Callbacks.Around = append(fsm.Desc.Callbacks.Around, transactionCb)
}
