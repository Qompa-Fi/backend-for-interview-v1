package main

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func main() {
	database, err := newDatabaseClient()
	if err != nil {
		panic(err)
	}

	defer database.Close()

	config := newConfig()

	router := echo.New()
	router.IPExtractor = echo.ExtractIPDirect()

	router.Use(middleware.CORS(), middleware.Recover(), middleware.Logger())

	wsManager := NewWSManager(config)
	defer wsManager.Close()

	router.GET("/ws/tasks", getHandleListenTasks(wsManager))
	router.GET("/ws/messages", nil)

	router.GET("/tasks", gethandleGetTasks())

	router.Logger.SetLevel(log.DEBUG)

	for _, route := range router.Routes() {
		router.Logger.Debug(route.Method + " - " + route.Path)
	}

	router.Start(fmt.Sprintf(":%d", config.Port))
}

/*
	websocket max connections per workspace -> 10
	max websocket workspaces per api key -> 3 + default(1)
*/
