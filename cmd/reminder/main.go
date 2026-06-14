package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reminder/auth"
	"reminder/config"
	"reminder/endpoint"
	"reminder/models"
	"reminder/service"
	httptransport "reminder/transport/http"
	"reminder/worker"
	"syscall"
	"time"

	kitlog "github.com/go-kit/log"
	gormsqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	logger := kitlog.NewLogfmtLogger(os.Stdout)
	logger.Log("msg", "Hello, World")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	db, err := gorm.Open(gormsqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("database connection error: %v", err)
	}
	db.AutoMigrate(&models.Reminder{}, &models.User{})
	if err := service.SeedUser(db, "alice", "password"); err != nil {
		log.Fatalf("seed user: %v", err)
	}

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	svc := service.NewReminderService(db)
	reminderWorker := worker.New(svc, logger, worker.Config{
		PollInterval: cfg.WorkerPollInterval,
	})

	go reminderWorker.Run(workerCtx)

	tokenSvc := auth.NewTokenService(cfg.JWTSecret, 24*time.Hour)
	authSvc := service.NewAuthService(db, tokenSvc)

	endpoints := endpoint.MakeEndpoints(svc, logger, tokenSvc, authSvc)
	handler := httptransport.NewHTTPHandler(endpoints)
	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: handler,
	}

	go func() {
		log.Printf("server listening on %s", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}

	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	workerCancel()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
	log.Println("server stopped")
}
