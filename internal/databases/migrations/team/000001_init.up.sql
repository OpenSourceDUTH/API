CREATE TABLE team_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    surname TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    github_username TEXT,
    email_notifications BOOLEAN DEFAULT 0 NOT NULL,
    is_admin BOOLEAN DEFAULT 0 NOT NULL,
    is_active BOOLEAN DEFAULT 1 NOT NULL,
    alumni BOOLEAN DEFAULT 0 NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Roles table for normalizing the multiple TeamRoles []string
-- This prevents comma-separated values in text columns and follows 1st normal form.
CREATE TABLE team_member_roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    member_id INTEGER NOT NULL,
    role TEXT NOT NULL,
    FOREIGN KEY (member_id) REFERENCES team_members(id) ON DELETE CASCADE,
    UNIQUE (member_id, role)
);

-- This project is the monolithic backend API for the OpenSourceDUTH team. Access to open data compiled and provided by the OpenSourceDUTH University Team.
-- API Copyright (C) 2025 OpenSourceDUTH
--     This program is free software: you can redistribute it and/or modify
--     it under the terms of the GNU General Public License as published by
--     the Free Software Foundation, either version 3 of the License, or
--     (at your option) any later version.

--     This program is distributed in the hope that it will be useful,
--     but WITHOUT ANY WARRANTY; without even the implied warranty of
--     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
--     GNU General Public License for more details.

--     You should have received a copy of the GNU General Public License
--     along with this program.  If not, see <https://www.gnu.org/licenses/>.