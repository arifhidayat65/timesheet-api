# Timesheet API (Go + Gin + PostgreSQL)

Arsitektur: Clean Architecture (Domain, Usecase, Repository, Transport) + Response helper.

## Quick start (dev)

```bash
cp .env.example .env
# Edit DB_DSN jika perlu

go mod tidy
go run ./cmd/api
```

Database:
```bash
sudo -u postgres psql -c "CREATE DATABASE timesheetdb;"
sudo -u postgres psql -c "CREATE USER user WITH PASSWORD 'password';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE timesheetdb TO user;"
```
