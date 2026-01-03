package service

import (
	"xlpanel/internal/domain"
	"xlpanel/internal/infra"
)

type SupportService struct {
	tickets *infra.Repository[domain.Ticket]
	metrics *infra.MetricsRegistry
}

func NewSupportService(tickets *infra.Repository[domain.Ticket], metrics *infra.MetricsRegistry) *SupportService {
	return &SupportService{tickets: tickets, metrics: metrics}
}

func (s *SupportService) OpenTicket(ticket domain.Ticket) domain.Ticket {
	s.metrics.Increment("ticket_opened")
	return s.tickets.Add(ticket)
}

func (s *SupportService) ListTickets() []domain.Ticket {
	return s.tickets.List()
}
