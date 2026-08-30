package main

import (
	"context"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kitis-api/internal/config"
	"kitis-api/internal/database"
	apiTransport "kitis-api/internal/transport"

	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// load config
	conf, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// init logger
	var level slog.Level
	switch conf.Logger.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	var logFile io.Writer
	if conf.Logger.Target == "file" {
		logFile = &lumberjack.Logger{
			Filename:   conf.Logger.Filename,
			MaxSize:    conf.Logger.MaxFileSizeMB,
			MaxBackups: conf.Logger.MaxFileBackups,
			MaxAge:     conf.Logger.MaxFileAgeDays,
			Compress:   conf.Logger.FileCompress,
		}
	} else {
		logFile = os.Stdout
	}
	logger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	// setup server
	s := apiTransport.NewServer(conf, ctx)
	go func() {
		slog.Info("server started", "addr", conf.Server.Address, "port", conf.Server.Port)
		if err := s.ListenAndServe(); err == http.ErrServerClosed {
			slog.Info("server closed")
		} else if err != nil {
			slog.Error("server closed with error", "cause", err)
		}
	}()
	defer database.ClosePool()

	// wait until server shuts down
	<-ctx.Done()
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		slog.Error("failed to shutdown server gracefully", "cause", err)
	}
}
