package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hibiken/asynq"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/openhost/openhost/internal/core/domain"
	infraPlugin "github.com/openhost/openhost/internal/infrastructure/plugin"
	provisionerv1 "github.com/openhost/openhost/pkg/proto/provisioner/v1"
)

const (
	ServiceStatusActive = "active"
)

type Worker struct {
	db      *gorm.DB
	plugins *infraPlugin.PluginManager
	logger  hclog.Logger
}

func NewWorker(db *gorm.DB, plugins *infraPlugin.PluginManager, logger hclog.Logger) *Worker {
	if logger == nil {
		logger = hclog.New(&hclog.LoggerOptions{
			Name:  "task-worker",
			Level: hclog.Info,
		})
	}
	return &Worker{
		db:      db,
		plugins: plugins,
		logger:  logger,
	}
}

func DefaultRetryDelay(retryCount int, _ error, _ *asynq.Task) time.Duration {
	if retryCount <= 0 {
		return 0
	}
	maxDelay := 10 * time.Minute
	delay := time.Duration(1<<min(retryCount, 10)) * time.Second
	if delay > maxDelay {
		return maxDelay
	}
	return delay
}

func (w *Worker) ProcessTask(ctx context.Context, task *asynq.Task) error {
	switch task.Type() {
	case TypeProvision:
		return w.handleProvision(ctx, task)
	case TypeSuspend:
		return asynq.SkipRetry
	case TypeTerminate:
		return asynq.SkipRetry
	default:
		return asynq.SkipRetry
	}
}

func (w *Worker) handleProvision(ctx context.Context, task *asynq.Task) error {
	if w.db == nil {
		return errors.New("db is required")
	}
	if w.plugins == nil {
		return errors.New("plugin manager is required")
	}

	var payload TaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}

	service, err := w.loadService(ctx, payload.ServiceID)
	if err != nil {
		return err
	}

	moduleName := service.Product.ModuleName
	if moduleName == "" {
		return errors.New("service product module name is required")
	}

	conn, err := w.plugins.GetClient(moduleName)
	if err != nil {
		return err
	}

	client := provisionerv1.NewProvisionerServiceClient(conn)
	request := buildProvisionRequest(service)
	if _, err := client.CreateService(ctx, request); err != nil {
		if statusErr := status.Convert(err); statusErr != nil {
			w.logger.Error("provisioner request failed", "service_id", service.ID, "error", statusErr.Message())
		}
		return err
	}

	if err := w.db.Model(&domain.Service{}).
		Where("id = ?", service.ID).
		Update("status", ServiceStatusActive).Error; err != nil {
		return fmt.Errorf("update service status: %w", err)
	}

	return nil
}

func (w *Worker) loadService(ctx context.Context, serviceID uint64) (domain.Service, error) {
	var service domain.Service
	if err := w.db.WithContext(ctx).
		Preload("Product").
		Preload("IPAddress").
		First(&service, serviceID).Error; err != nil {
		return domain.Service{}, fmt.Errorf("load service: %w", err)
	}
	return service, nil
}

func buildProvisionRequest(service domain.Service) *provisionerv1.CreateServiceRequest {
	options := map[string]string{}
	for key, value := range service.ConfigSelection {
		options[key] = stringifyOptionValue(value)
	}
	if service.IPAddress != nil {
		options["ip_address"] = service.IPAddress.IP
		options["gateway"] = service.IPAddress.Gateway
		options["netmask"] = service.IPAddress.Netmask
	}
	if service.PluginConfig.Region != "" {
		options["region"] = service.PluginConfig.Region
	}
	if service.PluginConfig.NodeID != "" {
		options["node_id"] = service.PluginConfig.NodeID
	}
	for key, value := range service.PluginConfig.Values {
		options[key] = value
	}

	return &provisionerv1.CreateServiceRequest{
		ServiceId:  strconv.FormatUint(service.ID, 10),
		CustomerId: strconv.FormatUint(service.CustomerID, 10),
		PackageId:  strconv.FormatUint(service.ProductID, 10),
		Options:    options,
	}
}

func stringifyOptionValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	case bool:
		return strconv.FormatBool(typed)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case uint64:
		return strconv.FormatUint(typed, 10)
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		data, err := json.Marshal(typed)
		if err != nil {
			return ""
		}
		return string(data)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
