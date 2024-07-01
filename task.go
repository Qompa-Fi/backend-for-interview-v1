package main

import (
	"math/rand"
	"sync"
	"time"
)

type (
	Task struct {
		mu     sync.Mutex
		logger SafeBuffer

		ID   uint64   `json:"id"`
		Name string   `json:"name"`
		Type TaskType `json:"type"`

		duration  time.Duration
		intensity uint64
	}

	TaskType string
)

// Type tasks. For fictional purposes only.
const (
	TinyTask   TaskType = "gx.tiny"
	MicroTask  TaskType = "gx.micro"
	LargeTask  TaskType = "gx.large"
	MediumTask TaskType = "gx.medium"
	SmallTask  TaskType = "gx.small"
	HeavyTask  TaskType = "gx.heavy"
)

func GetTaskTypes() []TaskType {
	return []TaskType{
		TinyTask,
		MicroTask,
		SmallTask,
		MediumTask,
		LargeTask,
		HeavyTask,
	}
}

func (tt TaskType) RandomDuration() time.Duration {
	var initDuration time.Duration

	switch tt {
	case TinyTask:
		initDuration = time.Second * 3
	case MicroTask:
		initDuration = time.Second * 5
	case SmallTask:
		initDuration = time.Second * 10
	case MediumTask:
		initDuration = time.Second * 30
	case LargeTask:
		initDuration = time.Minute * 2
	case HeavyTask:
		initDuration = time.Minute * 5
	}

	factor := rand.Float64()

	return time.Duration(float64(initDuration) * factor)
}

func (TaskType) RandomIntensity() float64 {
	return rand.Float64()
}

func NewTask(id uint64, name string, taskType TaskType) *Task {
	return &Task{
		ID:   id,
		Name: name,
	}
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
