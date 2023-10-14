package PS_DA

import (
	"fmt"
	"log"

	PS_structs "github.com/lapayka/rsoi-2/privilege-service/structs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	db *gorm.DB
}

func New(host, user, db_name, password string) (*DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s password=%s", host, user, db_name, password)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("unable to connect database", err)
	}

	return &DB{db: db}, nil
}

func (db *DB) GetPrivilegeAndHistoryByUserName(username string) (PS_structs.Privilege_with_history, error) {
	p := PS_structs.Privilege{Username: username}

	tx := db.db.Begin()

	err := tx.First(&p).Error

	if err != nil {
		tx.Rollback()
		return PS_structs.Privilege_with_history{}, nil
	}

	transactions := PS_structs.Privileges_history{}

	err = db.db.Find(&transactions).Where("Privilege_id = ", p.ID).Error

	if err != nil {
		tx.Rollback()
		return PS_structs.Privilege_with_history{Privilege_info: p}, nil
	}

	tx.Commit()

	return PS_structs.Privilege_with_history{Privilege_info: p, History: transactions}, err
}

func (db *DB) CreateTicket(username string, price int64, is_paid_from_balance bool, privelege_item PS_structs.Privilege_history) error {
	privelege := PS_structs.Privilege{Username: username}

	tx := db.db.Begin()
	err := tx.First(&privelege).Error

	if err != nil {
		tx.Rollback()
		return err
	}

	privelege_item.BalanceDiff = 0
	privelege_item.PrivilegeID = privelege.ID
	if is_paid_from_balance {
		diff := price
		if price > privelege.Balance {
			diff = privelege.Balance
		}
		privelege.Balance -= diff
		privelege_item.BalanceDiff = diff

		err = tx.Save(&privelege).Error

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Create(&privelege_item).Error

	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}
