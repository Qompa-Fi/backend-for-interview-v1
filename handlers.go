package main

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

func gethandleGetTasks(tm *TaskManager) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, tm.GetTasks())
	}
}

func getHandleListenTasks(wsManager *WSManager, tm *TaskManager) echo.HandlerFunc {
	return func(c echo.Context) error {
		workspace, err := wsManager.GetWorkspace(c)
		if err != nil {
			return err
		}

		defer workspace.Close()

		closeConn, err := workspace.NewConnection(c)
		if err != nil {
			return err
		}

		defer closeConn()

		ticker := time.NewTicker(time.Second)
		ctx := c.Request().Context()

		for range ticker.C {
			select {
			case <-ctx.Done():
				return nil
			default:
				tasks := tm.GetTasks()

				workspace.WriteMessage(websocket.TextMessage, mustJSONEncode(tasks))
			}
		}

		return nil
	}
}
