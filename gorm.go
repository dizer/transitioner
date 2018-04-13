package transitioner

import (
	"github.com/jinzhu/gorm"
)

type GormIntegration struct {
	// Tx *gorm.DB
	Transactional bool
}

type GormModel interface {
	GetDB() *gorm.DB
	SetDB(db *gorm.DB)
}

// func (fsm *FSM) IntegrateGorm(db *gorm.DB) {
func (fsm *FSM) IntegrateGorm(config GormIntegration) {

	fsm.Gorm = config

	transactionCb := func(fsm *FSM, fn func(*FSM, error) (*FSM, error)) func(*FSM, error) (*FSM, error) {
		return func(fsm *FSM, err error) (*FSM, error) {
			if err != nil {
				return fsm, err
			}

			object := fsm.Object.(GormModel)

			oldTx := object.GetDB()
			tx := oldTx.Begin()

			if tx.Error != nil {
				return fsm, tx.Error
			}

			defer func() {
				if r := recover(); r != nil {
					tx.Rollback()
					panic(r)
				}

				tx.Commit()
				object.SetDB(oldTx)
			}()

			object.SetDB(tx)
			retFSM, retERR := fn(fsm, err)

			return retFSM, retERR
		}
	}

	if fsm.Gorm.Transactional {
		fsm.Desc.Callbacks.Around = append(fsm.Desc.Callbacks.Around, transactionCb)
	}
}
