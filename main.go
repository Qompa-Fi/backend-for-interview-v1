package main

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func main() {
	config := newConfig()

	wsManager := NewWSManager(config)
	defer wsManager.Close()

	router := echo.New()
	router.Logger.SetLevel(log.DEBUG)
	router.IPExtractor = echo.ExtractIPDirect()

	router.Use(middleware.CORS(), middleware.Recover(), middleware.Logger())
	router.GET("/ws/tasks", getHandleListenTasks(wsManager))
	router.GET("/ws/messages", nil)
	router.GET("/tasks", gethandleGetTasks())

	for _, route := range router.Routes() {
		router.Logger.Debug(route.Method + " - " + route.Path)
	}

	router.Start(fmt.Sprintf(":%d", config.Port))
}
