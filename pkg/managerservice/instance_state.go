package managerservice

import (
	"context"
	"crypto/rand"
	"neko-manager/pkg/instancerepo"
	"neko-manager/pkg/nekosupplier"
	"neko-manager/pkg/randutils"
	"net"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/service_utils/logger_utils"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

func (r *Service) reportInstance(
	ctx context.Context,
	instance *instancerepo.Instance,
	text string,
	withStats bool,
) error {
	var statsPtr *nekosupplier.Stats
	if withStats && instance.IP != "" {
		stats, err := r.nekosupplier.GetStats(ctx, instance.IP, instance.SessionAPIToken)
		if err == nil {
			statsPtr = &stats
		}
	}

	var msgText = instance.Repr(statsPtr)
	if text != "" {
		msgText = text + "\n\n" + msgText
	}

	msg := tgbotapi.NewMessage(instance.TGChatID, msgText)
	msg.ParseMode = tgbotapi.ModeHTML

	_, err := r.terx.Bot.Send(msg)
	if err != nil {
		return errors.Wrap(err, "send tg message")
	}

	return nil
}

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
		CloudFolderID:   r.cloudSupplier.FolderID,
	}

	ctx = logger_utils.WithValue(ctx, "instance_id", instance.ID)

	err := r.instanceRepo.SaveInstance(ctx, &instance)
	if err != nil {
		return instancerepo.Instance{}, errors.Wrap(err, "save instance")
	}

	go r.HandleInstance(logger_utils.NewLoggedCtx(), &instance)

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
	instanceID := instance.ID
	ctx = logger_utils.WithValue(ctx, "instance_id", instanceID)
	zerolog.Ctx(ctx).Info().
		Stringer("instance", instance.Status).
		Msg("instance.handling.started")

	var (
		err           error
		sleepDuration = 5 * time.Second
	)
	for {
		instance, err = r.instanceRepo.GetInstance(ctx, instanceID)
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
	case instancerepo.InstanceStatusRunning:
		return r.processRunning(ctx, instance)
	case instancerepo.InstanceStatusDeleting:
		return r.processDeleting(ctx, instance)
	case instancerepo.InstanceStatusDeleted:
		return nil
	default:
		panic("invalid instance status")
	}
}

func (r *Service) createInstance(ctx context.Context, instance *instancerepo.Instance) error {
	cloudInstanceID, err := r.cloudSupplier.ComputeCreate(
		ctx,
		instance.ID,
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

		if computeState.GetStatus() != compute.Instance_RUNNING || len(computeState.GetNetworkInterfaces()) == 0 {
			zerolog.Ctx(ctx).Info().
				Str("state", computeState.GetStatus().String()).
				Msg("instance.is.not.running")
			time.Sleep(10 * time.Second)

			continue
		}

		instance.CloudInstanceID = cloudInstanceID

		address := net.ParseIP(
			computeState.GetNetworkInterfaces()[0].GetPrimaryV4Address().GetOneToOneNat().GetAddress(),
		)
		if address == nil {
			continue
		}

		instance.IP = address.String()
		instance.Status = instancerepo.InstanceStatusStarted

		err = r.instanceRepo.SaveInstance(ctx, instance)
		if err != nil {
			return errors.Wrap(err, "save instance")
		}

		return r.reportInstance(ctx, instance, "Cloud instance created, but neko is not ready yet", false)
	}
}

func (r *Service) waitForNekoStart(ctx context.Context, instance *instancerepo.Instance) error {
	for {
		_, err := r.nekosupplier.GetStats(ctx, instance.IP, instance.SessionAPIToken)
		if err != nil {
			zerolog.Ctx(ctx).
				Info().
				Err(err).
				Msg("neko.not.ready")
			time.Sleep(10 * time.Second)

			continue
		}

		instance.Status = instancerepo.InstanceStatusRunning

		err = r.instanceRepo.SaveInstance(ctx, instance)
		if err != nil {
			return errors.Wrap(err, "save instance")
		}

		return r.reportInstance(ctx, instance, "Neko ready!!!", true)
	}
}

func (r *Service) requireDeletion(ctx context.Context, instance *instancerepo.Instance) (bool, error) {
	stats, err := r.nekosupplier.GetStats(ctx, instance.IP, instance.SessionAPIToken)
	if err != nil {
		return false, errors.Wrap(err, "neko get stats")
	}

	if stats.TotalUsers != 0 || stats.TotalAdmins != 0 {
		zerolog.Ctx(ctx).Info().
			Interface("stats", stats).
			Msg("neko.instance.using")

		return false, nil
	}

	const maxIdle = time.Minute * 10

	now := time.Now().UTC()

	if stats.LastUsageAt().Add(maxIdle).Before(now) {
		return true, nil
	}

	zerolog.Ctx(ctx).Info().
		Interface("stats", stats).
		Msg("neko.instance.no.users")

	return false, nil
}

func (r *Service) processRunning(ctx context.Context, instance *instancerepo.Instance) error {
	for {
		ok, err := r.requireDeletion(ctx, instance)
		if err != nil {
			return errors.Wrap(err, "require deletion")
		}

		if !ok {
			time.Sleep(10 * time.Second)
			continue
		}

		instance.Status = instancerepo.InstanceStatusDeleting

		err = r.instanceRepo.SaveInstance(ctx, instance)
		if err != nil {
			return errors.Wrap(err, "save instance")
		}

		zerolog.Ctx(ctx).
			Info().
			Msg("neko.instance.deleting")

		return r.reportInstance(ctx, instance, "Deleting instance because of no usage", true)
	}
}

func (r *Service) Delete(ctx context.Context, instanceID string) error {
	instance, err := r.instanceRepo.GetInstance(ctx, instanceID)
	if err != nil {
		return errors.Wrap(err, "get instance")
	}

	instance.Status = instancerepo.InstanceStatusDeleting

	err = r.instanceRepo.SaveInstance(ctx, instance)
	if err != nil {
		return errors.Wrap(err, "save instance")
	}

	zerolog.Ctx(ctx).
		Info().
		Msg("neko.instance.deleting")

	return r.reportInstance(ctx, instance, "Deleting instance by request", true)
}

func (r *Service) processDeleting(ctx context.Context, instance *instancerepo.Instance) error {
	err := r.cloudSupplier.ComputeDeleteWaited(ctx, instance.CloudInstanceID)
	if err != nil {
		return errors.Wrap(err, "compute delete")
	}

	instance.Status = instancerepo.InstanceStatusDeleted

	err = r.instanceRepo.SaveInstance(ctx, instance)
	if err != nil {
		return errors.Wrap(err, "save instance")
	}

	return r.reportInstance(ctx, instance, "Instance deleted", false)
}
