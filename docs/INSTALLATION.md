# OpenHost Installation Guide

This guide walks you through installing OpenHost with the built-in SQLite database (recommended for quick starts) or PostgreSQL.

## Requirements

- Go 1.22+ (Go 1.23 toolchain recommended)
- GCC toolchain (required for SQLite driver `go-sqlite3`)
- Optional: PostgreSQL 13+

## Build the server

```bash
make server
```

The binary is created at `./bin/server`.

## Run the server

```bash
./bin/server
```

Open your browser at:

```
http://localhost:8080/install
```

## Web installer steps

1. **Site Settings**: Set the site name and base URL.
2. **Admin Account**: Provide the administrator email and password (stored as a bcrypt hash).
3. **Database**:
   - **SQLite (Built-in)**: Default path is `./data/openhost.db`.
   - **PostgreSQL**: Enter host, port, user, password, and database name.

After submission, the installer will:

- Create the database (or connect to PostgreSQL).
- Run GORM migrations for the core domain tables.
- Write configuration to `config/openhost.json`.

## Reinstalling

To rerun the installer:

1. Stop the server.
2. Delete `config/openhost.json`.
3. Remove the SQLite database file (if used), e.g. `./data/openhost.db`.
4. Start the server again and revisit `/install`.

## Notes

- The configuration file is created with `0600` permissions.
- SQLite data files are stored under `./data` by default.
- For PostgreSQL, ensure the database user has permission to create tables.
