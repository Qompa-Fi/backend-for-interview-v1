package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
)

type createTaskDto struct {
	Name string   `json:"name"`
	Type TaskType `json:"type"`
}

func getIndexHandler(router *echo.Echo) echo.HandlerFunc {
	return func(c echo.Context) error {
		txt := ` .              +   .                .   . .     .  .
                   .                    .       .     *
  .       *                        . . . .  .   .  + .
             You Are Here             .   .  +  . . .
.                 |             .  .   .    .    . .
                  |           .     .     . +.    +  .
                 \|/            .       .   . .
        . .       V          .    * . . .  .  +   .
           +      .           .   .      +
                            .       . +  .+. .
  .                      .     . + .  . .     .      .
           .      .    .     . .   . . .        ! /
      *             .    . .  +    .  .       - O -
          .     .    .  +   . .  *  .       . / |
               . + .  .  .  .. +  .
.      .  .  .  *   .  *  . +..  .            *
 .      .   . .   .   .   . .  +   .    .            +



Other places:
 - https://github.com/project-7pmwvjf9/backend-for-interview-v1
 - https://github.com/project-7pmwvjf9/frontend-interview-v1
`
		txt += "\nAvailable endpoints:\n"

		maxMethodLen := 0

		for _, route := range router.Routes() {
			if len(route.Method) > maxMethodLen {
				maxMethodLen = len(route.Method)
			}
		}

		for _, route := range router.Routes() {
			if route.Path == "/" {
				continue
			}

			txt += fmt.Sprintf("  %-*s - %s\n", maxMethodLen, route.Method, route.Path)
		}

		return c.String(http.StatusOK, txt)
	}
}

func getWSTasksHandler(m *Manager) echo.HandlerFunc {
	return func(c echo.Context) error {
		workspace, err := m.GetWorkspace(c)
		if err != nil {
			return err
		}

		if err := workspace.WSSubscribeToTasks(c); err != nil {
			return err
		}

		return nil
	}
}

func getGetTasksHandler(m *Manager) echo.HandlerFunc {
	return func(c echo.Context) error {
		workspace, err := m.GetWorkspace(c)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, echo.Map{
			"tasks": workspace.GetTasks(),
		})
	}
}

func getCreateTaskHandler(m *Manager) echo.HandlerFunc {
	return func(c echo.Context) error {
		workspace, err := m.GetWorkspace(c)
		if err != nil {
			return err
		}

		var dto createTaskDto

		if err := c.Bind(&dto); err != nil {
			c.Logger().Error(err)

			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		} else if dto.Name == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "name is required")
		} else if !lo.Contains(GetTaskTypes(), dto.Type) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid task type")
		}

		task := workspace.AddTask(dto.Name, dto.Type)

		return c.JSON(http.StatusOK, echo.Map{
			"task": task,
		})
	}
}

func getDeleteTaskHandler(m *Manager) echo.HandlerFunc {
	return func(c echo.Context) error {
		workspace, err := m.GetWorkspace(c)
		if err != nil {
			return err
		}

		rawId := c.Param("id")

		id, err := strconv.ParseUint(rawId, 10, strconv.IntSize)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid task id")
		}

		if err := workspace.DeleteTask(id); err != nil {
			if errors.Is(err, ErrTaskNotFound) {
				msg := "the task could not be found, it is either being dispatched or has already been dispatched"

				return echo.NewHTTPError(http.StatusNotFound, msg)
			}

			c.Logger().Error(err)

			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}

		return c.NoContent(http.StatusNoContent)
	}
}

func getFlushTasksHandler(m *Manager) echo.HandlerFunc {
	return func(c echo.Context) error {
		workspace, err := m.GetWorkspace(c)
		if err != nil {
			return err
		}

		if err := workspace.FlushTasks(); err != nil {
			c.Logger().Error(err)

			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}

		return c.NoContent(http.StatusNoContent)
	}
}
