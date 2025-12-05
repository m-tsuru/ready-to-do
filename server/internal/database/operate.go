package database

import (
	"gorm.io/gorm"
)

func Init(db *gorm.DB) error {
	// DuckDB の初期設定
	if err := InitDuckDB(db); err != nil {
		return err
	}

	// カスタムマイグレーション（DuckDB互換）
	if err := createTables(db); err != nil {
		return err
	}

	return nil
}

func createTables(db *gorm.DB) error {
	// users テーブル
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			mail TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}

	// user_sessions テーブル
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS user_sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			token TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NOT NULL
		)
	`).Error; err != nil {
		return err
	}

	// user_sessions のインデックス
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(token)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at)`)

	// tasks テーブル
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			user_id TEXT NOT NULL,
			description TEXT,
			related_url TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}

	// tasks のインデックス
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id)`)

	// parent_tasks テーブル
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS parent_tasks (
			id TEXT PRIMARY KEY,
			task_id TEXT NOT NULL,
			parent_id TEXT NOT NULL
		)
	`).Error; err != nil {
		return err
	}

	// parent_tasks のインデックス
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_parent_tasks_task_id ON parent_tasks(task_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_parent_tasks_parent_id ON parent_tasks(parent_id)`)

	// task_states テーブル
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS task_states (
			id TEXT PRIMARY KEY,
			task_id TEXT NOT NULL,
			state TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}

	// task_states のインデックス
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_task_states_task_id ON task_states(task_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_task_states_state ON task_states(state)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_task_states_created_at ON task_states(created_at)`)

	// task_state_change_logs テーブル
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS task_state_change_logs (
			id TEXT PRIMARY KEY,
			task_id TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			state TEXT NOT NULL
		)
	`).Error; err != nil {
		return err
	}

	// task_state_change_logs のインデックス
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_task_state_change_logs_task_id ON task_state_change_logs(task_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_task_state_change_logs_created_at ON task_state_change_logs(created_at)`)

	// task_running_logs テーブル
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS task_running_logs (
			id TEXT PRIMARY KEY,
			task_id TEXT NOT NULL,
			duration BIGINT NOT NULL
		)
	`).Error; err != nil {
		return err
	}

	// task_running_logs のインデックス
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_task_running_logs_task_id ON task_running_logs(task_id)`)

	return nil
}
