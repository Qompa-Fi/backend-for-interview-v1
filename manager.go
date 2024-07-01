package main

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
)

type Manager struct {
	mu        sync.RWMutex
	wsClients map[string]*WSClient
	config    *Config
}

var rxWorkspaceId = regexp.MustCompile(`^[A-z\-\_\d]+$`)

var ErrNoConnectionsInWorkspace = errors.New("no connections in workspace")

func NewManager(config *Config) *Manager {
	return &Manager{
		wsClients: make(map[string]*WSClient),
		config:    config,
	}
}

func (m *Manager) GetWorkspace(c echo.Context) (*Workspace, error) {
	apiKey, workspaceId, err := m.getApiKeyAndWorkspaceId(c)
	if err != nil {
		return nil, err
	}

	m.mu.RLock()
	client, ok := m.wsClients[apiKey]
	m.mu.RUnlock()

	if !ok {
		client = newWSClient(m.config)

		m.mu.Lock()
		m.wsClients[apiKey] = client
		m.mu.Unlock()
	}

	if c.IsWebSocket() {
		if client.GetNumberOfWorkspaces() >= m.config.MaxWorkspaces {
			return nil, echo.NewHTTPError(http.StatusTooManyRequests, "too many workspaces")
		}
	}

	workspace, err := client.GetWorkspace(c, apiKey, workspaceId)
	if err != nil {
		return nil, err
	}

	return workspace, nil
}

func (m *Manager) Close() {
	m.mu.Lock()
	for _, client := range m.wsClients {
		client.Close()
	}
	m.mu.Unlock()
}

func (m *Manager) getApiKeyAndWorkspaceId(c echo.Context) (apiKey string, workspaceId string, err error) {
	apiKey = strings.TrimSpace(c.QueryParam("api_key"))
	if !lo.Contains(m.config.APIKeys, apiKey) {
		return "", "", echo.ErrForbidden
	}

	workspaceId = strings.TrimSpace(c.QueryParam("workspace_id"))
	if workspaceId == "" {
		workspaceId = "default"
	}

	if !rxWorkspaceId.MatchString(workspaceId) {
		return "", "", echo.NewHTTPError(http.StatusBadRequest, "invalid workspace id")
	}

	return apiKey, workspaceId, nil
}
