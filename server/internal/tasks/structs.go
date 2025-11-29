package tasks

import "time"

type Task struct {
	Id          string    `db:"id"`
	Name        string    `db:"name"`
	UserId      string    `db:"user_id"` // User を埋め込まず参照にする
	Description string    `db:"description"`
	State       string    `db:"state"` // 例: "waiting","ready","running"...
	CreatedAt   time.Time `db:"created_at"`
}

type TaskState struct {
	Id        string    `db:"id"`
	TaskId    string    `db:"task_id"`
	State     string    `db:"state"`
	CreatedAt time.Time `db:"created_at"`
}

type TaskRunningLog struct {
	Id       string `db:"id"`
	TaskId   string `db:"task_id"`
	Duration int64  `db:"duration"`
}
