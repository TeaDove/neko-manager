package managerservice

import (
	"context"
	"neko-manager/pkg/instancerepo"

	"github.com/pkg/errors"

	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/service_utils/logger_utils"
	"github.com/teris-io/shortid"
)

type Service struct {
	instanceRepo *instancerepo.Repo
}

func New(instanceRepo *instancerepo.Repo) *Service {
	return &Service{instanceRepo: instanceRepo}
}

func (r *Service) RequestInstance(ctx context.Context) error {
	instance := instancerepo.Instance{
		ID:     shortid.MustGenerate(),
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

	return nil
}
