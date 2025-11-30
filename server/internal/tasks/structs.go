package tasks

import "time"

type Task struct {
	Id          string    `db:"id"`
	Name        string    `db:"name"`
	UserId      string    `db:"user_id"`
	Description string    `db:"description"`
	RelatedUrl  string    `db:"related_url"`
	CreatedAt   time.Time `db:"created_at"`
}

type ParentTask struct {
	Id       string `db:"id"`
	TaskId   string `db:"task_id"`
	ParentId string `db:"parent_id"`
}

type TaskState struct {
	Id        string    `db:"id"`
	TaskId    string    `db:"task_id"`
	State     string    `db:"state"` //タスク状態は "waiting", "running", "done" のいずれか
	CreatedAt time.Time `db:"created_at"`
}

type TaskRunningLog struct {
	Id       string `db:"id"`
	TaskId   string `db:"task_id"`
	Duration int64  `db:"duration"`
}
