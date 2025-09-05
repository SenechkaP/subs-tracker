package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SenechkaP/subs-tracker/configs"
	"github.com/SenechkaP/subs-tracker/internal/migrations"
	"github.com/SenechkaP/subs-tracker/internal/subscription"
	"github.com/SenechkaP/subs-tracker/pkg/db"
	"github.com/SenechkaP/subs-tracker/pkg/middleware"
)

func App(envPath string) http.Handler {
	conf := configs.LoadConfig(envPath)
	database := db.NewDb(conf)
	if err := migrations.RunMigrations(database); err != nil {
		log.Fatalf("migrate failed: %v", err)
	}

	router := http.NewServeMux()

	subscriptionRepository := subscription.NewSubscriptionRepository(database)

	subscription.NewSubscriptionHandler(router, &subscription.SubscriptionHandlerDeps{
		Repository: subscriptionRepository,
	})

	return middleware.Logging(router)
}

func main() {
	server := &http.Server{
		Addr:    ":8081",
		Handler: App(".env"),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("server shutdown failed:%+v", err)
	}
}
