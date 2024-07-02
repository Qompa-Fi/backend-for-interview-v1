package main

import (
	"fmt"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func main() {
	config := newConfig()

	manager := NewManager(config)
	defer manager.Close()

	router := echo.New()
	router.Logger.SetLevel(log.DEBUG)
	router.IPExtractor = echo.ExtractIPFromRealIPHeader()

	router.Use(middleware.CORS(), middleware.Recover(), middleware.Logger())

	router.GET("/", getIndexHandler(router))
	router.GET("/ws/tasks", getWSTasksHandler(manager))
	router.GET("/tasks", getGetTasksHandler(manager))
	router.POST("/tasks", getCreateTaskHandler(manager))
	router.DELETE("/tasks/:id", getDeleteTaskHandler(manager))
	router.POST("/tasks/flush", getFlushTasksHandler(manager))

	for _, route := range router.Routes() {
		fmt.Fprintf(os.Stderr, "%s %s - %s\n", time.Now().Local().Format(time.RFC3339), route.Method, route.Path)
	}

	router.Start(fmt.Sprintf(":%d", config.Port))
}
