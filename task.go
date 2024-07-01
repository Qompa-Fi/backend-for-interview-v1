package main

import (
	"fmt"
	"math/rand"
	"time"
)

type (
	TaskStatus string
	TaskType   string
)

const (
	StatusRunning   TaskStatus = "running"
	StatusCancelled TaskStatus = "cancelled"
	StatusQueued    TaskStatus = "queued"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
	// StatusPaused    TaskStatus = "paused"
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

func (tt TaskType) randomDuration() time.Duration {
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

func (TaskType) randomIntensity() float64 {
	return rand.Float64()
}

type Task struct {
	ID     uint64     `json:"id"`
	Name   string     `json:"name"`
	Type   TaskType   `json:"type"`
	Status TaskStatus `json:"status"`
}

func NewTask(id uint64, name string, taskType TaskType) *Task {
	return &Task{
		ID:     id,
		Name:   name,
		Type:   taskType,
		Status: StatusQueued,
	}
}

func (t *Task) GetDuration() time.Duration {
	return t.Type.randomDuration()
}

func (t *Task) GetIntensity() float64 {
	return t.Type.randomIntensity()
}

func (t *Task) SetStatus(status TaskStatus) {
	t.Status = status
}

func (t *Task) Start() {
	t.SetStatus(StatusRunning)

	d := t.GetDuration()
	fmt.Println(d)
	time.Sleep(d)

	t.SetStatus(StatusCompleted)
}
