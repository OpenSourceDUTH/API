# OpenSourceDUTH API

Running the migration command to create the SQLite database
```bash
go run cmd/migrate/main.go -path=schedule
```

Compiling the project
```bash
go build -o bin/api cmd/api/main.go
```

Running the API server
```bash
# In development mode
go tools air run
# Or directly
go run cmd/api/main.go   
# The binary
./bin/api
```


---
- - - 

## 🤝 Contributing
This project is part of our unified suite of apps for the students of Democritus University of Thrace, these apps are intended to help students with their university life. One of the main reasons why we do open-source is so that people can build upon and expand on our work on their own terms, so, questions, contributions and feature requests are more than welcome.

For our API documentation visit [opensource.cs.duth.gr/docs/](https://opensource.cs.duth.gr/docs/getting-started), for our contribution Guidelines visit our ["Contributing"]() page.

---
- - - 

### Dependencies
```json
"github.com/google/uuid"
"go-sqlite3"	
"github.com/golang-migrate/migrate/v4"
"github.com/golang-migrate/migrate/v4/database/sqlite3"
"github.com/golang-migrate/migrate/v4/source/file"

// Tools
"github.com/air-verse/air@latest"
```
```json
// TOOLS
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```
### License
`GNU General Public License v3.0`



<!--
This project is the monolithic backend API for the OpenSourceDUTH team. Access to open data compiled and provided by the OpenSourceDUTH University Team.
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
-->