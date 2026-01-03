package service

import (
	"time"

	"xlpanel/internal/infra"
)

type AutomationJob struct {
	Name     string    `json:"name"`
	Schedule string    `json:"schedule"`
	LastRun  time.Time `json:"last_run,omitempty"`
}

type AutomationService struct {
	metrics *infra.MetricsRegistry
	jobs    []AutomationJob
}

func NewAutomationService(metrics *infra.MetricsRegistry) *AutomationService {
	return &AutomationService{metrics: metrics, jobs: []AutomationJob{}}
}

func (s *AutomationService) RegisterJob(job AutomationJob) AutomationJob {
	s.jobs = append(s.jobs, job)
	s.metrics.Increment("automation_registered")
	return job
}

func (s *AutomationService) ListJobs() []AutomationJob {
	return append([]AutomationJob{}, s.jobs...)
}
