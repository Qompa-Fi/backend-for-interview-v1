package main

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type Workspace struct {
	mu          sync.RWMutex
	connections map[*websocket.Conn]struct{}

	inLoop bool
	tm     *TaskManager
}

func newWorkspace(config *Config) *Workspace {
	return &Workspace{
		connections: make(map[*websocket.Conn]struct{}),
		tm:          newTaskManager(config),
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

func (w *Workspace) WSSubscribeToTasks(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	w.mu.Lock()
	w.connections[conn] = struct{}{}
	w.mu.Unlock()

	return nil
}

func (w *Workspace) GetTasks() []*Task {
	return w.tm.GetTasks()
}

func (w *Workspace) AddTask(name string, taskType TaskType) *Task {
	return w.tm.AddTask(name, taskType)
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
