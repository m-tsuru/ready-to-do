package database

import (
	"github.com/m-tsuru/ready-to-do/internal/tasks"
	"github.com/m-tsuru/ready-to-do/internal/users"
	"gorm.io/gorm"
)

func Init(db *gorm.DB) error {
	err := db.AutoMigrate(
		&tasks.Task{},
		&tasks.TaskState{},
		&tasks.TaskRunningLog{},
		&users.User{},
		&users.UserSession{},
	)
	if err != nil {
		return err
	}
	return nil
}
