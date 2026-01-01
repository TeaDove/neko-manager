package managerservice

import (
	"context"
	"neko-manager/pkg/instancerepo"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (r *Service) restartOnErrRequired(instance *instancerepo.Instance) bool {
	return false
	// TODO fix this, now it's too unpredictable

	now := time.Now()

	if now.Sub(instance.UpdatedAt) < r.restartOnErrDuration {
		return false
	}

	if instance.LastHealthOk != nil {
		if now.Sub(*instance.LastHealthOk) > r.restartOnErrDuration {
			return true
		}
	} else {
		if now.Sub(instance.CreatedAt) > r.restartOnErrDuration {
			return true
		}
	}

	return false
}

func (r *Service) processRestarting(ctx context.Context, instance *instancerepo.Instance) (time.Duration, error) {
	if !r.restartOnErrRequired(instance) {
		return 10 * time.Second, nil
	}

	_, err := r.cloudSupplier.ComputeGet(ctx, instance.CloudName())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			instance.CloudInstanceID = nil
			instance.IP = nil

			instance.Status = instancerepo.InstanceStatusCreating

			return 0, r.saveAndReportInstance(ctx, instance, "Recreating VM", false)
		}

		return 0, errors.Wrap(err, "compute get")
	}

	err = r.cloudSupplier.ComputeRestartWaited(ctx, *instance.CloudInstanceID)
	if err != nil {
		return 0, errors.Wrap(err, "compute restart waited")
	}

	return 0, r.saveAndReportInstance(ctx, instance, "Restarting VM", false)
}
