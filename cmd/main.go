// @title           Subscription API
// @version         1.0
// @description REST-сервис для подписок пользователей
// @host            localhost:8080
// @BasePath        /

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"effective_mobile/config"
	_ "effective_mobile/docs"
	"effective_mobile/internal/db"
	"effective_mobile/internal/handler"
	"effective_mobile/internal/service"
	"effective_mobile/internal/storage"
)

func main() {
	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelFunc()

	cfg, err := config.GetConfig("config.yaml")
	if err != nil {
		slog.Error("Can't read config.yaml", slog.Any("error", err))
		return
	}

	dbConn, err := db.ConnectDB(ctx, cfg.Database.DSN)
	if err != nil {
		slog.Error("failed to connect", slog.Any("err", err))
		return
	}
	defer dbConn.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	store := storage.NewSubscriptionStorage(dbConn, logger)
	svc := service.NewSubscriptionService(store, logger)
	h := handler.NewSubscriptionHandler(svc, logger)

	router := handler.RegisterRoutes(h)

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), router)
		if err != nil {
			slog.Error("Can't start service:", slog.Any("error", err))
		}
	}()

	slog.Info("Starting HTTP server on", slog.String("port", strconv.Itoa(cfg.Server.Port)))

	<-ctx.Done()
	slog.Info("Got shutdown signal, exit program")
}
