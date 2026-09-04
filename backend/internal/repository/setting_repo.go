package repository

import (
	"database/sql"
)

type SettingRepo struct {
	db *sql.DB
}

func NewSettingRepo(db *sql.DB) *SettingRepo {
	return &SettingRepo{db: db}
}

// GetAllSettings retrieves all settings as a map
func (r *SettingRepo) GetAllSettings() (map[string]string, error) {
	rows, err := r.db.Query("SELECT key, value FROM app_settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err == nil {
			settings[key] = value
		}
	}
	return settings, nil
}

// GetSetting retrieves a specific setting by key
func (r *SettingRepo) GetSetting(key string) (string, error) {
	var value string
	err := r.db.QueryRow("SELECT value FROM app_settings WHERE key = $1", key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // return empty if not found
		}
		return "", err
	}
	return value, nil
}

// UpdateSettings updates multiple settings
func (r *SettingRepo) UpdateSettings(settings map[string]string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO app_settings (key, value, updated_at) 
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET 
		value=EXCLUDED.value, 
		updated_at=CURRENT_TIMESTAMP
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for k, v := range settings {
		if _, err := stmt.Exec(k, v); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
