package http

import (
	"context"
	"errors"
	"fmt"
	"kitis-api/internal/config"
	"kitis-api/internal/database"
	"kitis-api/internal/parser"
	api_http "kitis-api/internal/transport/http"
	api_ws "kitis-api/internal/transport/websocket"
	"kitis-api/internal/watcher"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"syscall"
	"time"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		next.ServeHTTP(w, r)
	})
}

func NewServer(conf *config.Config, ctx context.Context) *http.Server {
	mux := http.NewServeMux()

	// Подключение к postgresql
	db_conn_uri := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", conf.Database.Username, conf.Database.Password, conf.Database.Host, conf.Database.Port, conf.Database.Database)
	err := database.InitPool(ctx, db_conn_uri)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// init parser
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 50,
		IdleConnTimeout:     90 * time.Second,
	}
	cl := &http.Client{
		Transport: transport,
		Timeout:   conf.Parser.Timeout,
	}
	p := parser.NewParser(cl, &conf.Parser)
	err = p.UpdateSources(ctx)
	if err != nil {
		if errors.Is(err, syscall.ECONNRESET) {
			slog.Warn("skipping sources update")
		}
	}

	// init watcher
	w := watcher.NewWatcher(&conf.Watcher, p)

	// init handlers
	http_handler := api_http.NewHTTPHandler(conf, p, w)
	ws_handler := api_ws.NewWebSocketHandler(conf, w)
	ws_hub := api_ws.NewWebSocketHub(w)
	go ws_hub.Run(ctx)

	// register HTTP methods

	// v1
	// bells
	mux.HandleFunc("GET /v1/bells", http_handler.GetBells)
	// sources
	mux.HandleFunc("GET /v1/sources", http_handler.GetSources)
	mux.HandleFunc("GET /v1/sources/{source_type}", http_handler.GetSourcesByType)
	// schedule
	mux.HandleFunc("GET /v1/schedule/{source_type}/{source}", http_handler.GetSchedule)
	mux.HandleFunc("GET /v1/schedule/{source_type}/{source}/versions", http_handler.GetScheduleVersions)
	mux.HandleFunc("GET /v1/schedule/{source_type}/{source}/{timestamp}", http_handler.GetScheduleByTimestamp)
	mux.HandleFunc("GET /v1/schedule/db/{id}", http_handler.GetScheduleByID)
	// records
	mux.HandleFunc("GET /v1/records/{source_type}/{source}", http_handler.GetRecords)
	// journal
	mux.HandleFunc("GET /v1/journal/{id}", http_handler.GetJournalById)
	mux.HandleFunc("GET /v1/journal/{source}/{subgroup}/{lesson}", http_handler.GetJournal)

	// register websocket methods
	mux.HandleFunc("/v1/updates", func(w http.ResponseWriter, r *http.Request) {
		ws_handler.Updates(ctx, ws_hub, w, r)
	})

	// debug methods
	mux.HandleFunc("/v1/debug/ws/testupdate", func(w http.ResponseWriter, r *http.Request) {
		if !conf.Server.Debug {
			w.WriteHeader(404)
			return
		}
		ws_handler.TestNotification(ctx, w, r)
		w.WriteHeader(200)
	})

	// register not found method
	mux.HandleFunc("/", api_http.DefaultHandler)

	// start watcher
	go w.Start(ctx)

	return &http.Server{
		Addr:         ":" + strconv.Itoa(conf.Server.Port),
		Handler:      enableCORS(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}
