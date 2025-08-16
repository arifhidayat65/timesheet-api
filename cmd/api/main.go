package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"

	"timesheet-api/internal/config"
	appdb "timesheet-api/internal/db"
	"timesheet-api/internal/repository/postgres"
	"timesheet-api/internal/resp"
	transport "timesheet-api/internal/transport/http"
	"timesheet-api/internal/usecase"
	"timesheet-api/pkg/middleware"
)

func main() {
	cfg := config.Load()

	dbx := openPG(cfg.DB_DSN)
	defer dbx.Close()

	if err := appdb.Migrate(dbx); err != nil {
		log.Fatal(err)
	}

	repo := postgres.NewTimesheetRepoPG(dbx)
	svc := usecase.NewTimesheetService(repo)
	h := transport.NewTimesheetHandler(svc)

	r := gin.Default()
	r.Use(middleware.RequestID())
	r.Use(middleware.RecoveryJSON())

	r.GET("/health", func(c *gin.Context) {
		if err := dbx.Ping(); err != nil {
			resp.ServiceUnavailable(c, "DB down")
			return
		}
		resp.OK(c, gin.H{"status": "ok"}, "Healthy")
	})

	h.Register(r)

	log.Printf("listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}

func openPG(dsn string) *sql.DB {
	if dsn == "" {
		dsn = os.Getenv("DB_DSN")
	}
	if dsn == "" {
		log.Fatal("missing DB_DSN")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(`SET TIME ZONE 'Asia/Jakarta'`); err != nil {
		log.Printf("warn: set timezone: %v", err)
	}
	return db
}
