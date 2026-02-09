package schedule

import (
	"database/sql"
	"fmt"
	"time"
)

type Repository struct {
	db *sql.DB
}

// NewRepository creates a new schedule repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateFood adds a new food item to the database
func (r *Repository) CreateFood(name string) error {
	_, err := r.db.Exec("INSERT INTO foods (name) VALUES (?)", name)
	return err
}

// CreateVersion adds a new schedule version to the database
// TODO: Add validation for date formats
func (r *Repository) CreateVersion(start, end string, active bool) (int64, error) {
	res, err := r.db.Exec("INSERT INTO schedule_versions (starting_date, ending_date, is_current) VALUES (?, ?, ?)", start, end, active)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// CreateScheduleItem adds a new schedule item to the database with associated dishes. What day, week and meal type is this dish []int for.
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

// CreateAnnouncement adds a new announcement to the database
func (r *Repository) CreateAnnouncement(annType, content, start, end string, isCurrent bool) (int64, error) {
	res, err := r.db.Exec("INSERT INTO announcements (type, content, starting_date, ending_date, is_current) VALUES (?, ?, ?, ?, ?)", annType, content, start, end, isCurrent)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) GetDateSchedule(date string) (*DateSchedule, error) {
	var result DateSchedule

	// Avoid nil slices in JSON response
	result.Lunch = []Food{}
	result.Dinner = []Food{}

	var startingDateStr string
	var versionID int
	query := `SELECT id, starting_date FROM schedule_versions 
              WHERE ? >= starting_date AND (? <= ending_date OR ending_date IS NULL OR ending_date = '') 
              LIMIT 1`

	err := r.db.QueryRow(query, date, date).Scan(&versionID, &startingDateStr)
	if err != nil {
		return nil, err
	}
	// Trim time part if exists
	if len(startingDateStr) > 10 {
		startingDateStr = startingDateStr[:10]
	}

	start, err := time.Parse("2006-01-02", startingDateStr)
	if err != nil {
		return nil, err
	}
	target, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, err
	}

	daysDiff := int(target.Sub(start).Hours() / 24)
	if daysDiff < 0 {
		return nil, fmt.Errorf("We do not have a schedule for the requested date")
	}

	weekNum := ((daysDiff / 7) % 4) + 1
	dayNum := (daysDiff % 7) + 1

	rows, err := r.db.Query(`
        SELECT f.id, f.name, s.meal_type 
        FROM foods f
        JOIN schedule_dishes sd ON f.id = sd.food_id
        JOIN schedule s ON s.id = sd.schedule_id
        WHERE s.version_id = ? AND s.week_number = ? AND s.day_number = ?`, versionID, weekNum, dayNum)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var f Food
		var mealType string
		rows.Scan(&f.ID, &f.Name, &mealType)

		if mealType == "lunch" {
			result.Lunch = append(result.Lunch, f)
		} else {
			result.Dinner = append(result.Dinner, f)
		}
	}

	return &result, nil
}

// func (r *Repository) GetCurrentSchedule() {
// 	var result []CurrentSchedule
// 	var scheduleVersion ScheduleVersion

// 	err := r.db.QueryRow("SELECT id, starting_date, ending_date, is_current FROM schedule_versions WHERE is_current = 1 LIMIT 1").
// 		Scan(&scheduleVersion.ID, &scheduleVersion.StartingDate, &scheduleVersion.EndingDate, &scheduleVersion.IsCurrent)
// 	if err != nil {
// 		return
// 	}

// }

// func (r *Repository) GetAnnouncements(annType string) {

// }

//   This project is the monolithic backend API for the OpenSourceDUTH team. Access to open data compiled and provided by the OpenSourceDUTH University Team.
//   API Copyright (C) 2025 OpenSourceDUTH
//       This program is free software: you can redistribute it and/or modify
//       it under the terms of the GNU General Public License as published by
//       the Free Software Foundation, either version 3 of the License, or
//       (at your option) any later version.

//       This program is distributed in the hope that it will be useful,
//       but WITHOUT ANY WARRANTY; without even the implied warranty of
//       MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//       GNU General Public License for more details.

//       You should have received a copy of the GNU General Public License
//       along with this program.  If not, see <https://www.gnu.org/licenses/>.
