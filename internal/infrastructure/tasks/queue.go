package tasks

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const (
	TypeProvision = "openhost:provision"
	TypeSuspend   = "openhost:suspend"
	TypeTerminate = "openhost:terminate"
)

type TaskPayload struct {
	ServiceID uint64 `json:"service_id"`
}

func NewProvisionTask(serviceID uint64) (*asynq.Task, error) {
	return newTask(TypeProvision, TaskPayload{ServiceID: serviceID})
}

func NewSuspendTask(serviceID uint64) (*asynq.Task, error) {
	return newTask(TypeSuspend, TaskPayload{ServiceID: serviceID})
}

func NewTerminateTask(serviceID uint64) (*asynq.Task, error) {
	return newTask(TypeTerminate, TaskPayload{ServiceID: serviceID})
}

func newTask(taskType string, payload TaskPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(taskType, data), nil
}
