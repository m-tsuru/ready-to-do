package tasks

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CalculateRunningTime はタスクのrunning状態の合計時間を計算（秒単位）
func CalculateRunningTime(db *gorm.DB, taskId string) (int64, error) {
	var logs []TaskStateChangeLog

	// タスクの全ての状態変更ログを時系列順に取得
	if err := db.Where("task_id = ?", taskId).Order("created_at ASC").Find(&logs).Error; err != nil {
		return 0, err
	}

	if len(logs) == 0 {
		return 0, nil
	}

	var totalRunningSeconds int64 = 0
	var runningStartTime *time.Time

	for _, log := range logs {
		if log.State == "running" {
			// running状態になった時刻を記録
			runningStartTime = &log.CreatedAt
		} else if runningStartTime != nil {
			// running以外の状態になった場合、running期間を計算
			duration := log.CreatedAt.Sub(*runningStartTime)
			totalRunningSeconds += int64(duration.Seconds())
			runningStartTime = nil
		}
	}

	// 最後がrunning状態のまま終わっている場合、現在時刻までの時間を計算
	if runningStartTime != nil {
		duration := time.Since(*runningStartTime)
		totalRunningSeconds += int64(duration.Seconds())
	}

	return totalRunningSeconds, nil
}

// GetParentDependency は指定されたタスクの親タスクをすべて取得
func GetParentDependency(db *gorm.DB, t *Task) (*[]Task, error) {
	var parentTasks []Task

	// ParentTask テーブルから親タスクのIDを取得
	var parentRelations []ParentTask
	if err := db.Where("task_id = ?", t.Id).Find(&parentRelations).Error; err != nil {
		return nil, err
	}

	// 親タスクがない場合は空の配列を返す
	if len(parentRelations) == 0 {
		return &[]Task{}, nil
	}

	// 親タスクのIDリストを作成
	parentIds := make([]string, len(parentRelations))
	for i, rel := range parentRelations {
		parentIds[i] = rel.ParentId
	}

	// 親タスクを取得
	if err := db.Where("id IN ?", parentIds).Find(&parentTasks).Error; err != nil {
		return nil, err
	}

	return &parentTasks, nil
}

// GetChildDependency は指定されたタスクの子タスクをすべて取得
func GetChildDependency(db *gorm.DB, t *Task) (*[]Task, error) {
	var childTasks []Task

	// ParentTask テーブルから子タスクのIDを取得
	var childRelations []ParentTask
	if err := db.Where("parent_id = ?", t.Id).Find(&childRelations).Error; err != nil {
		return nil, err
	}

	// 子タスクがない場合は空の配列を返す
	if len(childRelations) == 0 {
		return &[]Task{}, nil
	}

	// 子タスクのIDリストを作成
	childIds := make([]string, len(childRelations))
	for i, rel := range childRelations {
		childIds[i] = rel.TaskId
	}

	// 子タスクを取得
	if err := db.Where("id IN ?", childIds).Find(&childTasks).Error; err != nil {
		return nil, err
	}

	return &childTasks, nil
}

// GetTaskById は ID でタスクを取得
func GetTaskById(db *gorm.DB, Id string) (*Task, error) {
	var task Task
	if err := db.Where("id = ?", Id).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// GetTasksByUserId はユーザーIDでタスクを取得
func GetTasksByUserId(db *gorm.DB, userId string) (*[]Task, error) {
	var tasks []Task
	if err := db.Where("user_id = ?", userId).Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return &tasks, nil
}

// CreateTask はタスクを作成
func CreateTask(db *gorm.DB, userId string, name string, description string, relatedUrl string, parentTaskIds []string) (*Task, error) {
	// タスクを作成。タスク ID は UUID
	task := &Task{
		Id:          uuid.New().String(),
		Name:        name,
		UserId:      userId,
		Description: description,
		RelatedUrl:  relatedUrl,
		CreatedAt:   time.Now(),
	}

	if err := db.Create(task).Error; err != nil {
		return nil, err
	}

	// 親タスクの指示があればその ID を ParentTask に登録
	if len(parentTaskIds) > 0 {
		for _, parentId := range parentTaskIds {
			parentTask := &ParentTask{
				Id:       uuid.New().String(),
				TaskId:   task.Id,
				ParentId: parentId,
			}
			if err := db.Create(parentTask).Error; err != nil {
				return nil, err
			}
		}
	}

	// タスク状態 (TaskState) を "waiting" で作成
	taskState := &TaskState{
		Id:        uuid.New().String(),
		TaskId:    task.Id,
		State:     "waiting",
		CreatedAt: time.Now(),
	}

	if err := db.Create(taskState).Error; err != nil {
		return nil, err
	}

	// 初期状態の変更ログを記録
	changeLog := &TaskStateChangeLog{
		Id:        uuid.New().String(),
		TaskId:    task.Id,
		State:     "waiting",
		CreatedAt: time.Now(),
	}

	if err := db.Create(changeLog).Error; err != nil {
		return nil, err
	}

	return task, nil
}

// GetCurrentState は現在のタスク状態を取得
func (t Task) GetCurrentState(db *gorm.DB) (*TaskState, error) {
	var state TaskState
	if err := db.Where("task_id = ?", t.Id).Order("created_at DESC").First(&state).Error; err != nil {
		return nil, err
	}
	return &state, nil
}

// IsReady は waiting のタスクの時、親タスクがすべて done なら true を返却
func (t Task) IsReady(db *gorm.DB) (bool, error) {
	// 現在の状態を取得
	currentState, err := t.GetCurrentState(db)
	if err != nil {
		return false, err
	}

	// waiting 状態でない場合は false
	if currentState.State != "waiting" {
		return false, nil
	}

	// 親タスクを取得
	parentTasks, err := GetParentDependency(db, &t)
	if err != nil {
		return false, err
	}

	// 親タスクがない場合は ready
	if len(*parentTasks) == 0 {
		return true, nil
	}

	// すべての親タスクが done かチェック
	for _, parent := range *parentTasks {
		parentState, err := parent.GetCurrentState(db)
		if err != nil {
			return false, err
		}
		if parentState.State != "done" {
			return false, nil
		}
	}

	return true, nil
}

// MakeRunning は ready のタスクの時、状態を running に変更
func (t Task) MakeRunning(db *gorm.DB) (bool, error) {
	// ready 状態かチェック
	isReady, err := t.IsReady(db)
	if err != nil {
		return false, err
	}
	if !isReady {
		return false, nil
	}

	// 新しい状態を作成
	newState := &TaskState{
		Id:        uuid.New().String(),
		TaskId:    t.Id,
		State:     "running",
		CreatedAt: time.Now(),
	}

	if err := db.Create(newState).Error; err != nil {
		return false, err
	}

	// 状態変更ログを記録
	changeLog := &TaskStateChangeLog{
		Id:        uuid.New().String(),
		TaskId:    t.Id,
		State:     "running",
		CreatedAt: time.Now(),
	}

	if err := db.Create(changeLog).Error; err != nil {
		return false, err
	}

	return true, nil
}

// MakeWaiting は running のタスクの時、状態を waiting に変更
func (t Task) MakeWaiting(db *gorm.DB) (bool, error) {
	// 現在の状態を取得
	currentState, err := t.GetCurrentState(db)
	if err != nil {
		return false, err
	}

	// running 状態でない場合は false
	if currentState.State != "running" {
		return false, nil
	}

	// 新しい状態を作成
	newState := &TaskState{
		Id:        uuid.New().String(),
		TaskId:    t.Id,
		State:     "waiting",
		CreatedAt: time.Now(),
	}

	if err := db.Create(newState).Error; err != nil {
		return false, err
	}

	// 状態変更ログを記録
	changeLog := &TaskStateChangeLog{
		Id:        uuid.New().String(),
		TaskId:    t.Id,
		State:     "waiting",
		CreatedAt: time.Now(),
	}

	if err := db.Create(changeLog).Error; err != nil {
		return false, err
	}

	return true, nil
}

// MakeDone は waiting, running のタスクの時、状態を done に変更
func (t Task) MakeDone(db *gorm.DB) (bool, error) {
	// 現在の状態を取得
	currentState, err := t.GetCurrentState(db)
	if err != nil {
		return false, err
	}

	// waiting または running 状態でない場合は false
	if currentState.State != "waiting" && currentState.State != "running" {
		return false, nil
	}

	// 新しい状態を作成
	newState := &TaskState{
		Id:        uuid.New().String(),
		TaskId:    t.Id,
		State:     "done",
		CreatedAt: time.Now(),
	}

	if err := db.Create(newState).Error; err != nil {
		return false, err
	}

	// 状態変更ログを記録
	changeLog := &TaskStateChangeLog{
		Id:        uuid.New().String(),
		TaskId:    t.Id,
		State:     "done",
		CreatedAt: time.Now(),
	}

	if err := db.Create(changeLog).Error; err != nil {
		return false, err
	}

	return true, nil
}
