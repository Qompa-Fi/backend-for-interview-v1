package main

import (
	"fmt"
	"net"
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
		fmt.Printf("c.QueryParam(\"workspace_id\"): %v\n", c.QueryParam("workspace_id"))
		fmt.Printf("c.QueryParam(\"api_key\"): %v\n", c.QueryParam("api_key"))

		ws, err := wsManager.GetConnection(c)
		if err != nil {
			return err
		}

		defer ws.Close()

		ticker := time.NewTicker(time.Second)
		ctx := c.Request().Context()

		for range ticker.C {
			select {
			case <-ctx.Done():
				return nil
			default:
				tasks := tm.GetTasks()

				err := ws.WriteMessage(websocket.TextMessage, mustJSONEncode(tasks))
				if err != nil {
					if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) {
						return nil
					}

					c.Logger().Error(err)

					if _, ok := err.(*net.OpError); ok {
						return nil
					}

					continue
				}
			}
		}

		return nil
	}
}
