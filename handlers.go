package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
)

type createTaskDto struct {
	Name string   `json:"name"`
	Type TaskType `json:"type"`
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
