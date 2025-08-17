package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"

	"timesheet-api/internal/config"
	appdb "timesheet-api/internal/db"
	"timesheet-api/internal/repository/postgres"   // ← BENAR (tanpa alias 'http')
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

	repo := postgres.NewTimesheetRepoPG(dbx) // ⬅️ panggil lewat nama paket "postgres"
	svc := usecase.NewTimesheetService(repo)
	h := transport.NewTimesheetHandler(svc)

	r := gin.Default()

	// Trusted proxies aman utk lokal & docker
	if err := r.SetTrustedProxies([]string{
		"127.0.0.1", "::1",
		"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16",
	}); err != nil {
		log.Fatal(err)
	}
	r.ForwardedByClientIP = true

	r.Use(middleware.RequestID(), middleware.RecoveryJSON())

	r.GET("/health", func(c *gin.Context) {
		if err := dbx.Ping(); err != nil {
			resp.ServiceUnavailable(c, "DB down")
			return
		}
		resp.OK(c, gin.H{"status": "ok"}, "Healthy")
	})

	h.Register(r)
	for _, ri := range r.Routes() {
	log.Printf("[ROUTE] %s %s -> %s", ri.Method, ri.Path, ri.Handler)
	}

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
	var db *sql.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = sql.Open("pgx", dsn)
		if err == nil && db.Ping() == nil {
			_, _ = db.Exec(`SET TIME ZONE 'Asia/Jakarta'`)
			return db
		}
		time.Sleep(time.Duration(500*(i+1)) * time.Millisecond)
	}
	log.Fatalf("failed to connect DB: %v", err)
	return nil
}
