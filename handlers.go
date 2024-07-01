package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type createTaskDto struct {
	Name string `json:"name" validate:"required"`
}

func getGetTasksHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, nil)
	}
}

func getWSTasksHandler(m *Manager) echo.HandlerFunc {
	return func(c echo.Context) error {
		workspace, err := m.GetWorkspace(c)
		if err != nil {
			return err
		}

		if err := workspace.Subscribe(c); err != nil {
			return err
		}

		return nil
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
		}

		task := workspace.AddTask(dto.Name)

		return c.JSON(http.StatusOK, echo.Map{
			"task": task,
		})
	}
}
