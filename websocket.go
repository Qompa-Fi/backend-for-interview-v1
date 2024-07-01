package main

import (
	"errors"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/samber/lo"
)

type Manager struct {
	mu        sync.RWMutex
	wsClients map[string]*WSClient
	config    *Config
}

var (
	rxWorkspaceId = regexp.MustCompile(`^[A-z\-\_\d]+$`)
	upgrader      websocket.Upgrader
)

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

	if !workspace.inLoop {
		go workspace.serveConnectionsLoop()
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

	inLoop bool
	tm     *TaskManager
}

func newWorkspace() *Workspace {
	return &Workspace{
		connections: make(map[*websocket.Conn]struct{}),
		tm:          newTaskManager(),
	}
}

func (w *Workspace) serveConnectionsLoop() {
	defer w.Close()

	log.Info("starting new loop...")

	ticker := time.NewTicker(time.Second)
	w.inLoop = true

	for range ticker.C {
		tasks := w.tm.GetTasks()

		err := w.WriteMessage(websocket.TextMessage, mustJSONEncode(tasks))
		if err != nil {
			if errors.Is(err, ErrNoConnectionsInWorkspace) {
				log.Debug("no connections in workspace...")
			} else {
				log.Error(err)
			}

			log.Info("closing loop...")
			w.inLoop = false

			break
		}
	}
}

func (w *Workspace) GetConnectionCount() uint64 {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return uint64(len(w.connections))
}

func (w *Workspace) WriteMessage(messageType int, data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.connections) == 0 {
		return ErrNoConnectionsInWorkspace
	}

	for conn := range w.connections {
		err := conn.WriteMessage(messageType, data)
		if err != nil {
			if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) {
				delete(w.connections, conn)
				conn.Close()
			}

			log.Error(err)

			if _, ok := err.(*net.OpError); ok {
				delete(w.connections, conn)
				conn.Close()
			}
		}
	}

	if len(w.connections) == 0 {
		return ErrNoConnectionsInWorkspace
	}

	return nil
}

func (w *Workspace) Subscribe(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	w.mu.Lock()
	w.connections[conn] = struct{}{}
	w.mu.Unlock()

	return nil
}

func (w *Workspace) AddTask(name string) *Task {
	return w.tm.AddTask(name)
}

func (w *Workspace) Close() {
	w.mu.Lock()

	for conn := range w.connections {
		delete(w.connections, conn)
		if err := conn.Close(); err != nil {
			log.Error(err)
		}
	}

	w.mu.Unlock()
}
