package main

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/samber/lo"
)

type (
	TaskManager struct {
		mu     sync.RWMutex
		nextId uint64

		config *Config

		dispatchedTasks   []*Task
		undispatchedTasks []*Task
	}
)

func newTaskManager(config *Config) *TaskManager {
	tm := &TaskManager{
		dispatchedTasks:   make([]*Task, 0),
		undispatchedTasks: make([]*Task, 0),
		config:            config,
		nextId:            1,
	}

	go tm.dispatchTasksLoop()

	return tm
}

func (m *TaskManager) AddTask(name string, taskType TaskType) *Task {
	m.mu.Lock()
	defer m.mu.Unlock()

	task := m.newTask(name, taskType)

	m.undispatchedTasks = append(m.undispatchedTasks, task)

	return task
}

func (m *TaskManager) GetTasks() []*Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return append(m.dispatchedTasks, m.undispatchedTasks...)
}

func (m *TaskManager) dispatchTasksLoop() {
	runningChan := make(chan struct{}, m.config.MaxRunningWorkspaceTasks)

	for {
		task, ok := m.getNextTask()
		if !ok {
			time.Sleep(time.Second)
			continue
		}

		runningChan <- struct{}{}

		go func() {
			defer func() {
				<-runningChan
			}()

			task.Start()
		}()

		time.Sleep(time.Millisecond * 100)
	}
}

func (m *TaskManager) getNextTask() (*Task, bool) {
	m.mu.RLock()
	task, index, ok := lo.FindIndexOf(m.undispatchedTasks, func(t *Task) bool {
		return t.Status == StatusQueued
	})
	m.mu.RUnlock()
	if !ok {
		return nil, false
	}

	m.mu.Lock()
	m.undispatchedTasks = append(m.undispatchedTasks[:index], m.undispatchedTasks[index+1:]...)
	m.dispatchedTasks = append(m.dispatchedTasks, task)
	m.mu.Unlock()

	return task, true
}

func (m *TaskManager) newTask(name string, taskType TaskType) *Task {
	id := m.nextId
	atomic.AddUint64(&m.nextId, 1)

	return &Task{
		ID:     id,
		Name:   name,
		Type:   taskType,
		Status: StatusQueued,
	}
}
