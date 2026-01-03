package service

import (
	"xlpanel/internal/domain"
	"xlpanel/internal/infra"
)

type HostingService struct {
	vps     *infra.Repository[domain.VPSInstance]
	ips     *infra.Repository[domain.IPAllocation]
	metrics *infra.MetricsRegistry
}

func NewHostingService(
	vps *infra.Repository[domain.VPSInstance],
	ips *infra.Repository[domain.IPAllocation],
	metrics *infra.MetricsRegistry,
) *HostingService {
	return &HostingService{vps: vps, ips: ips, metrics: metrics}
}

func (s *HostingService) ProvisionVPS(instance domain.VPSInstance) domain.VPSInstance {
	s.metrics.Increment("vps_provisioned")
	return s.vps.Add(instance)
}

func (s *HostingService) AllocateIP(allocation domain.IPAllocation) domain.IPAllocation {
	s.metrics.Increment("ip_allocated")
	return s.ips.Add(allocation)
}

func (s *HostingService) ListVPS() []domain.VPSInstance {
	return s.vps.List()
}

func (s *HostingService) ListIPs() []domain.IPAllocation {
	return s.ips.List()
}
