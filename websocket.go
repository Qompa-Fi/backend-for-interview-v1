package main

import (
	"net/http"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/samber/lo"
)

type WSManager struct {
	mu      sync.RWMutex
	clients map[string]*WSClient

	config *Config
}

var (
	rxWorkspaceId = regexp.MustCompile(`^[A-z\-\_\d]+$`)
	upgrader      websocket.Upgrader
)

func NewWSManager(config *Config) *WSManager {
	return &WSManager{
		clients: make(map[string]*WSClient),
		config:  config,
	}
}

func (m *WSManager) GetConnection(c echo.Context) (*websocket.Conn, error) {
	apiKey, workspaceId, err := m.getApiKeyAndWorkspaceId(c)
	if err != nil {
		return nil, err
	}

	m.mu.RLock()
	client, ok := m.clients[apiKey]
	m.mu.RUnlock()

	if !ok {
		client = newWSClient(m.config)

		m.mu.Lock()
		m.clients[apiKey] = client
		m.mu.Unlock()
	}

	if client.GetNumberOfWorkspaces() >= m.config.MaxWorkspaces {
		return nil, echo.NewHTTPError(http.StatusTooManyRequests, "too many workspaces")
	}

	conn, err := client.GetWorkspaceConnection(c, workspaceId)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (m *WSManager) Close() {
	m.mu.Lock()
	for _, client := range m.clients {
		client.Close()
	}
	m.mu.Unlock()
}

func (m *WSManager) getApiKeyAndWorkspaceId(c echo.Context) (apiKey string, workspaceId string, err error) {
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

type (
	WSClient struct {
		mu         sync.RWMutex
		workspaces map[string]Workspace
		config     *Config
	}

	Workspace struct {
		count      uint64
		connection *websocket.Conn
	}
)

func newWSClient(config *Config) *WSClient {
	return &WSClient{
		workspaces: make(map[string]Workspace),
		config:     config,
	}
}

func (wsc *WSClient) GetNumberOfWorkspaces() uint64 {
	wsc.mu.RLock()
	defer wsc.mu.RUnlock()

	return uint64(len(wsc.workspaces))
}

func (wsc *WSClient) GetWorkspaceConnection(c echo.Context, workspaceId string) (*websocket.Conn, error) {
	wsc.mu.RLock()
	workspace, ok := wsc.workspaces[workspaceId]
	wsc.mu.RUnlock()

	if workspace.count >= wsc.config.MaxWorkspaceConnections {
		return nil, echo.NewHTTPError(http.StatusTooManyRequests, "too many connections")
	}

	atomic.AddUint64(&workspace.count, 1)

	if ok {
		return workspace.connection, nil
	}

	conn, err := wsc.createWorkspaceConnection(c, workspaceId)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (wsc *WSClient) createWorkspaceConnection(c echo.Context, workspaceId string) (*websocket.Conn, error) {
	wsc.mu.RLock()
	workspace, ok := wsc.workspaces[workspaceId]
	wsc.mu.RUnlock()

	if ok {
		return workspace.connection, nil
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return nil, err
	}

	wsc.mu.Lock()
	wsc.workspaces[workspaceId] = Workspace{connection: conn, count: 0}
	wsc.mu.Unlock()

	return conn, nil
}

func (wsc *WSClient) Close() {
	wsc.mu.Lock()
	for _, workspace := range wsc.workspaces {
		if err := workspace.connection.Close(); err != nil {
			log.Error(err)
		}
	}

	wsc.mu.Unlock()
}
