package managerservice

import (
	"context"
	"neko-manager/pkg/instancerepo"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func (r *Service) DeleteUnknown(ctx context.Context) error {
	instances, err := r.instanceRepo.ListActiveInstances(ctx)
	if err != nil {
		return errors.Wrap(err, "list active instances")
	}

	knownInstancesID := mapset.NewThreadUnsafeSet[string]()

	for _, instance := range instances {
		if instance.CloudInstanceID != nil {
			knownInstancesID.Add(*instance.CloudInstanceID)
		}
	}

	cloudInstances, err := r.cloudSupplier.ComputeList(ctx, instancerepo.NEKO)
	if err != nil {
		return errors.Wrap(err, "compute list")
	}

	for _, cloudInstance := range cloudInstances {
		if knownInstancesID.Contains(cloudInstance.GetId()) {
			continue
		}

		zerolog.Ctx(ctx).Info().
			Str("name", cloudInstance.GetName()).
			Str("id", cloudInstance.GetId()).
			Msg("deteting.unknown.instance")

		err = r.cloudSupplier.ComputeDeleteWaited(ctx, cloudInstance.GetId())
		if err != nil {
			return errors.Wrap(err, "delete instance")
		}
	}

	return nil
}
