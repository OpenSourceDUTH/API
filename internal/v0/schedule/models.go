package schedule

type Food struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ScheduleVersion struct {
	ID           int    `json:"id"`
	StartingDate string `json:"starting_date"`
	EndingDate   string `json:"ending_date"`
	IsCurrent    bool   `json:"is_current"`
}

type ScheduleItem struct {
	ID         int    `json:"id"`
	VersionID  int    `json:"version_id"`
	WeekNumber int    `json:"week_number"`
	DayNumber  int    `json:"day_number"`
	MealType   string `json:"meal_type"`
	DishIDs    []int  `json:"dish_ids"`
}
type Announcement struct {
	ID           int    `json:"id"`
	Type         string `json:"type"`
	Content      string `json:"content"`
	StartingDate string `json:"starting_date"`
	EndingDate   string `json:"ending_date"`
	IsCurrent    bool   `json:"is_current"`
}

type DateSchedule struct {
	Lunch  []Food `json:"lunch"`
	Dinner []Food `json:"dinner"`
}

type SemesterSchedule map[int]map[int]DateSchedule

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
