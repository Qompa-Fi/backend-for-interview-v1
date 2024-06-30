package main

import (
	"sync"
	"time"
)

type Task struct {
	mu     sync.Mutex
	logger SafeBuffer

	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

func (t *Task) Start() {
	t.logger.WriteString("Task started\n")
}

func (t *Task) Cancel() {
	t.logger.WriteString("Task cancelled\n")
}

func (t *Task) Log(message string) {
	t.logger.WriteString(`{"message": "` + message + `", "at": "` + time.Now().Local().String() + `"}\n`)
}
