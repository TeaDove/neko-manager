package managerservice

import (
	"context"
	"crypto/rand"
	"fmt"
	"neko-manager/pkg/instancerepo"
	"neko-manager/pkg/randutils"
	"net"
	"time"

	"github.com/pkg/errors"
	"github.com/teadove/teasutils/utils/test_utils"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/service_utils/logger_utils"
)

func (r *Service) RequestInstance(
	ctx context.Context,
	tgChatID int64,
	createdBy string,
) (instancerepo.Instance, error) {
	instance := instancerepo.Instance{
		ID:              randutils.RandomString(6),
		Status:          instancerepo.InstanceStatusCreating,
		CreatedBy:       createdBy,
		TGChatID:        tgChatID,
		SessionAPIToken: rand.Text(),
	}

	ctx = logger_utils.WithValue(ctx, "instance_id", instance.ID)

	err := r.instanceRepo.SaveInstance(ctx, &instance)
	if err != nil {
		return instancerepo.Instance{}, errors.Wrap(err, "save instance")
	}

	go r.HandleInstance(ctx, &instance)

	zerolog.Ctx(ctx).
		Info().
		Msg("neko.instance.creating")

	return instance, nil
}

func (r *Service) Reconciliation(ctx context.Context) error {
	instances, err := r.instanceRepo.ListActiveInstances(ctx)
	if err != nil {
		return errors.Wrap(err, "list active instances")
	}

	for _, instance := range instances {
		go r.HandleInstance(ctx, &instance)
	}

	return nil
}

func (r *Service) HandleInstance(ctx context.Context, instance *instancerepo.Instance) {
	ctx = logger_utils.WithValue(ctx, "instance_id", instance.ID)
	zerolog.Ctx(ctx).Info().
		Stringer("instance", instance.Status).
		Msg("instance.handling.started")

	var (
		err           error
		sleepDuration = 5 * time.Second
	)
	for {
		instance, err = r.instanceRepo.GetInstance(ctx, instance.ID)
		if err != nil {
			zerolog.Ctx(ctx).Error().Stack().Err(err).Msg("failed.to.get.instance")
			time.Sleep(sleepDuration)

			continue
		}

		if instance.Status == instancerepo.InstanceStatusDeleted {
			return
		}

		err = r.processInstanceStatus(ctx, instance)
		if err != nil {
			zerolog.Ctx(ctx).Error().Stack().Err(err).Msg("failed.to.process.instance.state")
			time.Sleep(sleepDuration)

			continue
		}
	}
}

func (r *Service) processInstanceStatus(ctx context.Context, instance *instancerepo.Instance) error {
	switch instance.Status {
	case instancerepo.InstanceStatusCreating:
		return r.createInstance(ctx, instance)
	case instancerepo.InstanceStatusStarted:
		return r.waitForNekoStart(ctx, instance)
	default:
		panic("invalid instance status")
	}
}

func (r *Service) createInstance(ctx context.Context, instance *instancerepo.Instance) error {
	cloudInstanceID, err := r.cloudSupplier.ComputeCreate(
		ctx,
		instance.CloudName(),
		instance.CreatedBy,
		instance.SessionAPIToken,
	)
	if err != nil {
		return errors.Wrap(err, "compute create")
	}

	for {
		computeState, err := r.cloudSupplier.ComputeGet(ctx, instance.CloudName())
		if err != nil {
			return errors.Wrap(err, "compute get")
		}

		if computeState.GetStatus() == compute.Instance_RUNNING && len(computeState.GetNetworkInterfaces()) != 0 {
			instance.CloudInstanceID = cloudInstanceID

			test_utils.Pprint(computeState.GetNetworkInterfaces())

			instance.IP = net.ParseIP(computeState.GetNetworkInterfaces()[0].GetPrimaryV4Address().GetAddress())
			if instance.IP == nil {
				continue
			}

			instance.Status = instancerepo.InstanceStatusStarted

			err = r.instanceRepo.SaveInstance(ctx, instance)
			if err != nil {
				return errors.Wrap(err, "save instance")
			}

			_, err = r.terx.Bot.Send(tgbotapi.NewMessage(instance.TGChatID,
				fmt.Sprintf("Cloud instance created: %s", instance.ID)),
			)
			if err != nil {
				return errors.Wrap(err, "send tg message")
			}

			return nil
		}

		time.Sleep(10 * time.Second)
	}
}

func (r *Service) waitForNekoStart(ctx context.Context, instance *instancerepo.Instance) error {
	for {
		err := r.nekosupplier.GetStats(ctx, instance.IP.String(), instance.SessionAPIToken)
		if err != nil {
			zerolog.Ctx(ctx).
				Info().
				Err(err).
				Msg("neko.get.stats")
			time.Sleep(10 * time.Second)
		}

		return nil
	}
}
