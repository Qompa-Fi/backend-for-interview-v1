package main

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	database, err := newDatabaseClient()
	if err != nil {
		panic(err)
	}

	defer database.Close()

	config := newConfig()

	tm := newTaskManager()

	tm.AddTask("nao nao")
	tm.AddTask("just another task")
	tm.AddTask("panzer")

	router := echo.New()
	router.IPExtractor = echo.ExtractIPDirect()

	router.Use(middleware.CORS(), middleware.Recover(), middleware.Logger())

	router.GET("/ws/tasks?workspace_id=dasdasd", nil)
	router.GET("/ws/messages?workspace_id=dasdasd", nil)

	router.GET("/tasks", gethandleGetTasks(tm))
	router.GET("/ws/tasks", getHandleListenTasks(tm))

	for _, route := range router.Routes() {
		router.Logger.Debug(route.Method + " - " + route.Path)
	}

	router.Start(fmt.Sprintf(":%d", config.Port))
}
