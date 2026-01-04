package main

import (
	"errors"
	"flag"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	path := flag.String("path", "schedule", "path to the database file")
	flag.Parse()

	m, err := migrate.New(
		"file://internal/databases/migrations/"+*path,
		"sqlite3://internal/databases/"+*path+".db",
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal(err)
	}
	log.Println("Database migration complete for the:", *path, "path")
}

/*
This project is the monolithic backend API for the OpenSourceDUTH team. Access to open data compiled and provided by the OpenSourceDUTH University Team as well as helper endpoints to integrate with our apps.
API Copyright (C) 2025 OpenSourceDUTH
    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
