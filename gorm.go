package transitioner

import (
	"github.com/jinzhu/gorm"
)

type GormIntegration struct {
	Tx *gorm.DB
	Transactional bool
}

// func (fsm *FSM) IntegrateGorm(db *gorm.DB) {
func (fsm *FSM) IntegrateGorm(config GormIntegration) {

	fsm.Gorm = config

	transactionCb := func(fsm *FSM, fn func(*FSM, error) (*FSM, error)) func(*FSM, error) (*FSM, error) {
		return func(fsm *FSM, err error) (*FSM, error) {
			if err != nil {
				return fsm, err
			}

			oldTx := fsm.Gorm.Tx
			tx := fsm.Gorm.Tx.Begin()

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

	if fsm.Gorm.Transactional {
		fsm.Desc.Callbacks.Around = append(fsm.Desc.Callbacks.Around, transactionCb)
	}
}
