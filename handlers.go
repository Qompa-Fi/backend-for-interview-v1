package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func gethandleGetTasks() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, nil)
	}
}

func getHandleListenTasks(wsManager *WSManager) echo.HandlerFunc {
	return func(c echo.Context) error {
		workspace, err := wsManager.GetWorkspace(c)
		if err != nil {
			return err
		}

		if err := workspace.Subscribe(c); err != nil {
			return err
		}

		return nil
	}
}
