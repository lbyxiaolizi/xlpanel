package provisionerv1

import (
	"context"

	"google.golang.org/grpc"
)

type CreateServiceRequest struct {
	ServiceId  string
	CustomerId string
	PackageId  string
	Options    map[string]string
}

type CreateServiceResponse struct {
	ExternalId string
	Message    string
}

type SuspendRequest struct {
	ServiceId string
	Reason    string
}

type SuspendResponse struct {
	Message string
}

type TerminateRequest struct {
	ServiceId string
	Reason    string
}

type TerminateResponse struct {
	Message string
}

type ChangePackageRequest struct {
	ServiceId        string
	CurrentPackageId string
	TargetPackageId  string
	Reason           string
}

type ChangePackageResponse struct {
	Message string
}

type PowerAction int32

const (
	PowerAction_POWER_ACTION_UNSPECIFIED PowerAction = 0
	PowerAction_POWER_ACTION_START       PowerAction = 1
	PowerAction_POWER_ACTION_STOP        PowerAction = 2
	PowerAction_POWER_ACTION_REBOOT      PowerAction = 3
)

type PowerControlRequest struct {
	ServiceId string
	Action    PowerAction
}

type PowerControlResponse struct {
	Message string
}

type GetUsageRequest struct {
	ServiceId string
}

type UsageMetric struct {
	Name  string
	Unit  string
	Value float64
}

type GetUsageResponse struct {
	Metrics []*UsageMetric
	Message string
}

type ProvisionerServiceClient interface {
	CreateService(ctx context.Context, in *CreateServiceRequest, opts ...grpc.CallOption) (*CreateServiceResponse, error)
	Suspend(ctx context.Context, in *SuspendRequest, opts ...grpc.CallOption) (*SuspendResponse, error)
	Terminate(ctx context.Context, in *TerminateRequest, opts ...grpc.CallOption) (*TerminateResponse, error)
	ChangePackage(ctx context.Context, in *ChangePackageRequest, opts ...grpc.CallOption) (*ChangePackageResponse, error)
	PowerControl(ctx context.Context, in *PowerControlRequest, opts ...grpc.CallOption) (*PowerControlResponse, error)
	GetUsage(ctx context.Context, in *GetUsageRequest, opts ...grpc.CallOption) (*GetUsageResponse, error)
}

type provisionerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewProvisionerServiceClient(cc grpc.ClientConnInterface) ProvisionerServiceClient {
	return &provisionerServiceClient{cc}
}

func (c *provisionerServiceClient) CreateService(ctx context.Context, in *CreateServiceRequest, opts ...grpc.CallOption) (*CreateServiceResponse, error) {
	out := new(CreateServiceResponse)
	err := c.cc.Invoke(ctx, "/openhost.plugin.provisioner.v1.ProvisionerService/CreateService", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *provisionerServiceClient) Suspend(ctx context.Context, in *SuspendRequest, opts ...grpc.CallOption) (*SuspendResponse, error) {
	out := new(SuspendResponse)
	err := c.cc.Invoke(ctx, "/openhost.plugin.provisioner.v1.ProvisionerService/Suspend", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *provisionerServiceClient) Terminate(ctx context.Context, in *TerminateRequest, opts ...grpc.CallOption) (*TerminateResponse, error) {
	out := new(TerminateResponse)
	err := c.cc.Invoke(ctx, "/openhost.plugin.provisioner.v1.ProvisionerService/Terminate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *provisionerServiceClient) ChangePackage(ctx context.Context, in *ChangePackageRequest, opts ...grpc.CallOption) (*ChangePackageResponse, error) {
	out := new(ChangePackageResponse)
	err := c.cc.Invoke(ctx, "/openhost.plugin.provisioner.v1.ProvisionerService/ChangePackage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *provisionerServiceClient) PowerControl(ctx context.Context, in *PowerControlRequest, opts ...grpc.CallOption) (*PowerControlResponse, error) {
	out := new(PowerControlResponse)
	err := c.cc.Invoke(ctx, "/openhost.plugin.provisioner.v1.ProvisionerService/PowerControl", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *provisionerServiceClient) GetUsage(ctx context.Context, in *GetUsageRequest, opts ...grpc.CallOption) (*GetUsageResponse, error) {
	out := new(GetUsageResponse)
	err := c.cc.Invoke(ctx, "/openhost.plugin.provisioner.v1.ProvisionerService/GetUsage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
