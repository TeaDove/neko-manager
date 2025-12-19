package managerservice

import (
	"context"
	"neko-manager/pkg/cloudsupplier"
	"neko-manager/pkg/instancerepo"
	"neko-manager/pkg/randutils"

	"github.com/pkg/errors"

	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/service_utils/logger_utils"
)

type Service struct {
	instanceRepo  *instancerepo.Repo
	cloudSupplier *cloudsupplier.Supplier
}

func New(instanceRepo *instancerepo.Repo, cloudSupplier *cloudsupplier.Supplier) *Service {
	return &Service{instanceRepo: instanceRepo, cloudSupplier: cloudSupplier}
}

func (r *Service) RequestInstance(ctx context.Context) error {
	instance := instancerepo.Instance{
		ID:     randutils.RandomString(6),
		Status: instancerepo.InstanceStatusCreating,
	}

	ctx = logger_utils.WithValue(ctx, "instance_id", instance.ID)

	err := r.instanceRepo.SaveInstance(ctx, &instance)
	if err != nil {
		return errors.Wrap(err, "save instance")
	}

	zerolog.Ctx(ctx).
		Info().
		Msg("instance.creating")

	err = r.cloudSupplier.ComputeCreate(ctx, instance.CloudName(), instance.CreatedBy)
	if err != nil {
		return errors.Wrap(err, "cloud supplier list")
	}

	return nil
}
