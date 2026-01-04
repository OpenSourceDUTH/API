package schedule

import (
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateFood(name string) error {
	_, err := r.db.Exec("INSERT INTO foods (name) VALUES (?)", name)
	return err
}

func (r *Repository) CreateVersion(start, end string, active bool) (int64, error) {
	res, err := r.db.Exec("INSERT INTO schedule_versions (starting_date, ending_date, is_current) VALUES (?, ?, ?)", start, end, active)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) CreateScheduleItem(versionID int, week, day int, mealType string, dishIDs []int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	// Defer a rollback in case anything fails.
	defer func() {
		_ = tx.Rollback()
	}()

	res, err := tx.Exec(`
		INSERT INTO schedule (version_id, week_number, day_number, meal_type) 
		VALUES (?, ?, ?, ?)`,
		versionID, week, day, mealType,
	)
	if err != nil {
		return err
	}

	scheduleID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO schedule_dishes (schedule_id, food_id) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, foodID := range dishIDs {
		if _, err := stmt.Exec(scheduleID, foodID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) CreateAnnouncement(annType, content, start, end string, isCurrent bool) (int64, error) {
	res, err := r.db.Exec("INSERT INTO announcements (type, content, starting_date, ending_date, is_current) VALUES (?, ?, ?, ?, ?)", annType, content, start, end, isCurrent)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
