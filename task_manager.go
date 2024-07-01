package main

import "sync"

type TaskManager struct {
	mu    sync.RWMutex
	tasks []*Task
}

func newTaskManager() *TaskManager {
	return &TaskManager{
		tasks: make([]*Task, 0),
	}
}

func (m *TaskManager) AddTask(name string) *Task {
	m.mu.Lock()
	defer m.mu.Unlock()

	task := &Task{
		ID:   uint64(len(m.tasks)) + 1,
		Name: name,
	}

	m.tasks = append(m.tasks, task)

	return task
}

func (m *TaskManager) GetTasks() []*Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.tasks
}

func (m *TaskManager) CancelTask(task *Task) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, t := range m.tasks {
		if t.ID == task.ID {
			m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
			return
		}
	}
}
