CREATE TABLE foods (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
);

CREATE TABLE schedule_versions(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    starting_date DATE NOT NULL,
    ending_date DATE,
    is_current BOOLEAN DEFAULT 0 NOT NULL
);

CREATE TABLE schedule(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version_id INTEGER NOT NULL,
    week_number INTEGER CHECK (week_number BETWEEN 1 AND 4),
    day_number INTEGER CHECK (day_number BETWEEN 1 AND 7),
    meal_type TEXT CHECK (meal_type IN ('lunch', 'dinner')),
    FOREIGN KEY (version_id) REFERENCES schedule_versions(id)
);

CREATE TABLE schedule_dishes(
    schedule_id INTEGER NOT NULL,
    food_id INTEGER NOT NULL,
    PRIMARY KEY (schedule_id, food_id),
    FOREIGN KEY (schedule_id) REFERENCES schedule(id) ON DELETE CASCADE,
    FOREIGN KEY (food_id) REFERENCES foods(id) ON DELETE CASCADE
);

CREATE TABLE announcements(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT CHECK (type in ('info','menu_change','holiday', 'emergency')),
    content TEXT NOT NULL,
    starting_date DATE NOT NULL,
    ending_date DATE,
    is_current BOOLEAN DEFAULT 0
);
