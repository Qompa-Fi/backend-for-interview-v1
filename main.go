package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func main() {
	now := time.Now()

	config := newConfig()

	manager := NewManager(config)
	defer manager.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	router := echo.New()
	router.Logger.SetLevel(log.DEBUG)
	router.IPExtractor = echo.ExtractIPFromRealIPHeader()

	router.Use(middleware.CORS(), middleware.Recover(), middleware.Logger())

	router.GET("/", getIndexHandler(router))
	router.GET("/stats", getServerStatsHandler(&now))
	router.GET("/ws/tasks", getWSTasksHandler(manager))
	router.GET("/tasks", getGetTasksHandler(manager))
	router.POST("/tasks", getCreateTaskHandler(manager))
	router.DELETE("/tasks/:id", getDeleteTaskHandler(manager))
	router.POST("/tasks/flush", getFlushTasksHandler(manager))

	for _, route := range router.Routes() {
		fmt.Fprintf(os.Stderr, "%s %s - %s\n", time.Now().Local().Format(time.RFC3339), route.Method, route.Path)
	}

	go func() {
		if err := router.Start(fmt.Sprintf(":%d", config.Port)); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Fatal(err)
			}
		}
	}()

	<-ctx.Done()

	if err := router.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Infof("uptime: %s", time.Since(now))
}
