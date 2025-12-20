package managerservice

import (
	"context"
	"neko-manager/pkg/cloudsupplier"
	"neko-manager/pkg/instancerepo"
	"neko-manager/pkg/randutils"
	"time"

	"github.com/pkg/errors"
	"github.com/teadove/teasutils/utils/test_utils"
	"github.com/teadove/terx/terx"

	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/service_utils/logger_utils"
)

type Service struct {
	instanceRepo  *instancerepo.Repo
	cloudSupplier *cloudsupplier.Supplier
	terx          *terx.Terx
}

func New(instanceRepo *instancerepo.Repo, cloudSupplier *cloudsupplier.Supplier, terx *terx.Terx) *Service {
	return &Service{instanceRepo: instanceRepo, cloudSupplier: cloudSupplier, terx: terx}
}

func (r *Service) RequestInstance(ctx context.Context, tgChatID int64, createdBy string) (instancerepo.Instance, error) {
	instance := instancerepo.Instance{
		ID:        randutils.RandomString(6),
		Status:    instancerepo.InstanceStatusCreating,
		CreatedBy: createdBy,
		TGChatID:  tgChatID,
	}

	ctx = logger_utils.WithValue(ctx, "instance_id", instance.ID)

	err := r.instanceRepo.SaveInstance(ctx, &instance)
	if err != nil {
		return instancerepo.Instance{}, errors.Wrap(err, "save instance")
	}

	zerolog.Ctx(ctx).
		Info().
		Msg("neko.instance.creating")

	//err = r.cloudSupplier.ComputeCreate(ctx, instance.CloudName(), instance.CreatedBy)
	//if err != nil {
	//	return instancerepo.Instance{}, errors.Wrap(err, "cloud supplier list")
	//}

	return instance, nil
}

func (r *Service) Reconciliation(ctx context.Context) {
	const period = time.Minute
	for {
		instances, err := r.instanceRepo.ListActiveInstances(ctx)
		if err != nil {
			zerolog.Ctx(ctx).
				Error().
				Stack().Err(err).
				Msg("failed.to.list.active.instances")

			time.Sleep(period)
			continue
		}

		test_utils.Pprint(instances)
		time.Sleep(period)

	}
}
