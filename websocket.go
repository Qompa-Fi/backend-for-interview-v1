package main

import (
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"

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

func (m *WSManager) GetWorkspace(c echo.Context) (*Workspace, error) {
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

	conn, err := client.GetWorkspace(c, apiKey, workspaceId)
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

type WSClient struct {
	mu         sync.RWMutex
	workspaces map[string]*Workspace
	config     *Config
}

func newWSClient(config *Config) *WSClient {
	return &WSClient{
		workspaces: make(map[string]*Workspace),
		config:     config,
	}
}

func (wsc *WSClient) GetNumberOfWorkspaces() uint64 {
	wsc.mu.RLock()
	defer wsc.mu.RUnlock()

	return uint64(len(wsc.workspaces))
}

func (wsc *WSClient) GetWorkspace(c echo.Context, apiKey, workspaceId string) (*Workspace, error) {
	wsc.mu.RLock()
	workspace, ok := wsc.workspaces[workspaceId]
	wsc.mu.RUnlock()

	if !ok {
		workspace = newWorkspace()

		log.Infof("new workspace '%s' created for client with api key '%s'", workspaceId, apiKey)

		wsc.mu.Lock()
		wsc.workspaces[workspaceId] = workspace
		wsc.mu.Unlock()
	} else {
		log.Infof("incoming connection to workspace '%s' created for client with api key '%s'", workspaceId, apiKey)
	}

	if workspace.GetConnectionCount() >= wsc.config.MaxWorkspaceConnections {
		return nil, echo.NewHTTPError(http.StatusTooManyRequests, "too many connections in workspace")
	}

	return workspace, nil
}

func (wsc *WSClient) Close() {
	wsc.mu.Lock()

	for _, workspace := range wsc.workspaces {
		workspace.Close()
	}

	wsc.mu.Unlock()
}

type Workspace struct {
	mu          sync.RWMutex
	connections map[*websocket.Conn]struct{}
}

func newWorkspace() *Workspace {
	return &Workspace{
		connections: make(map[*websocket.Conn]struct{}),
	}
}

func (w *Workspace) GetConnectionCount() uint64 {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return uint64(len(w.connections))
}

func (w *Workspace) WriteMessage(messageType int, data []byte) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	for conn := range w.connections {
		err := conn.WriteMessage(messageType, data)
		if err != nil {
			if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) {
				continue
			}

			log.Error(err)

			if _, ok := err.(*net.OpError); ok {
				continue
			}
		}
	}
}

func (w *Workspace) NewConnection(c echo.Context) (func(), error) {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return nil, err
	}

	w.mu.Lock()
	w.connections[conn] = struct{}{}
	w.mu.Unlock()

	return func() {
		w.mu.Lock()
		delete(w.connections, conn)
		w.mu.Unlock()

		if err := conn.Close(); err != nil {
			log.Error(err)
		}
	}, nil
}

func (w *Workspace) Close() {
	w.mu.Lock()

	for conn := range w.connections {
		if err := conn.Close(); err != nil {
			log.Error(err)
		}
	}

	w.mu.Unlock()
}
