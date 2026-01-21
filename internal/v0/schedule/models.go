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
