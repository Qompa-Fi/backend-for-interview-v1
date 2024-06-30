package main

import (
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

var upgrader websocket.Upgrader

func gethandleGetTasks(tm *TaskManager) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, tm.GetTasks())
	}
}

func getHandleListenTasks(tm *TaskManager) echo.HandlerFunc {
	return func(c echo.Context) error {
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			log.Error(err)

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
