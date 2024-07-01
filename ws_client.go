package main

import (
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

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
		workspace = newWorkspace(wsc.config)

		log.Infof("new workspace '%s' created for client with api key '%s'", workspaceId, apiKey)

		wsc.mu.Lock()
		wsc.workspaces[workspaceId] = workspace
		wsc.mu.Unlock()
	} else {
		if c.IsWebSocket() {
			log.Infof("incoming WS connection to workspace '%s' created for client with api key '%s'", workspaceId, apiKey)
		}
	}

	if c.IsWebSocket() && !workspace.inLoop {
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
