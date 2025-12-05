package tasks

import "time"

type Task struct {
	Id          string    `gorm:"column:id;primaryKey" db:"id" json:"id"`
	Name        string    `gorm:"column:name;not null" db:"name" json:"name"`
	UserId      string    `gorm:"column:user_id;not null;index" db:"user_id" json:"user_id"`
	Description string    `gorm:"column:description" db:"description" json:"description"`
	RelatedUrl  string    `gorm:"column:related_url" db:"related_url" json:"related_url"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" db:"created_at" json:"created_at"`
}

func (Task) TableName() string {
	return "tasks"
}

type ParentTask struct {
	Id       string `gorm:"column:id;primaryKey" db:"id"`
	TaskId   string `gorm:"column:task_id;not null;index" db:"task_id"`
	ParentId string `gorm:"column:parent_id;not null;index" db:"parent_id"`
}

func (ParentTask) TableName() string {
	return "parent_tasks"
}

type TaskState struct {
	Id        string    `gorm:"column:id;primaryKey" db:"id" json:"id"`
	TaskId    string    `gorm:"column:task_id;not null;index" db:"task_id" json:"task_id"`
	State     string    `gorm:"column:state;not null;index" db:"state" json:"state"` //タスク状態は "waiting", "running", "done" のいずれか
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index" db:"created_at" json:"created_at"`
}

func (TaskState) TableName() string {
	return "task_states"
}

type TaskStateChangeLog struct {
	Id        string    `gorm:"column:id;primaryKey" db:"id" json:"id"`
	TaskId    string    `gorm:"column:task_id;not null;index" db:"task_id" json:"task_id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" db:"created_at" json:"created_at"`
	State     string    `gorm:"column:state;not null" db:"state" json:"state"`
}

func (TaskStateChangeLog) TableName() string {
	return "task_state_change_logs"
}
