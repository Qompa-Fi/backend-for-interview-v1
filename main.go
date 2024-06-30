package main

import (
	"fmt"
	"net/http"

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

	router.GET("/tasks", func(c echo.Context) error {
		return c.JSON(http.StatusOK, tm.GetTasks())
	})

	router.Start(fmt.Sprintf(":%d", config.Port))
}
