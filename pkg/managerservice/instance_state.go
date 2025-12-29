package managerservice

import (
	"context"
	"crypto/rand"
	"neko-manager/pkg/instancerepo"
	"neko-manager/pkg/nekosupplier"
	"neko-manager/pkg/randutils"
	"net"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/service_utils/logger_utils"
	"github.com/teadove/teasutils/utils/time_utils"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

func (r *Service) RequestInstance(
	ctx context.Context,
	tgChatID int64,
	threadChatID int,
	createdBy string,
) (instancerepo.Instance, error) {
	instance := instancerepo.Instance{
		ID:              randutils.RandomString(6),
		Status:          instancerepo.InstanceStatusCreating,
		CreatedBy:       createdBy,
		TGChatID:        tgChatID,
		TGThreadChatID:  threadChatID,
		SessionAPIToken: rand.Text(),
		CloudFolderID:   r.cloudSupplier.FolderID,
	}
	if r.proxy.URL != "" {
		instance.ProxyURL = r.proxy.URL
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

	zerolog.Ctx(ctx).
		Info().
		Object("instance", instance).
		Msg("instance.handling.started")

	var (
		err error
	)

	const (
		sleepOnErrDuration = 5 * time.Second
	)

	for {
		instance, err = r.instanceRepo.GetInstance(ctx, instanceID)
		if err != nil {
			zerolog.Ctx(ctx).Error().Stack().Err(err).Msg("failed.to.get.instance")
			time.Sleep(sleepOnErrDuration)

			continue
		}

		if instance.Status == instancerepo.InstanceStatusDeleted {
			return
		}

		sleepDuration, err := r.processInstanceStatus(ctx, instance)
		if err != nil {
			zerolog.Ctx(ctx).
				Error().
				Stack().Err(err).
				Object("instance", instance).
				Msg("failed.to.process.instance.state")
			time.Sleep(sleepOnErrDuration)

			continue
		}

		if sleepDuration != 0 {
			zerolog.Ctx(ctx).
				Info().
				Object("instance", instance).
				Msg("sleeping.on.processing")
			time.Sleep(sleepDuration)
		}
	}
}

func (r *Service) processInstanceStatus(ctx context.Context, instance *instancerepo.Instance) (time.Duration, error) {
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
		return 0, nil
	default:
		panic("invalid instance status")
	}
}

func (r *Service) createInstance(ctx context.Context, instance *instancerepo.Instance) (time.Duration, error) {
	cloudInstanceID, err := r.cloudSupplier.ComputeCreateWaited(
		ctx,
		instance.ID,
		instance.CloudName(),
		instance.CreatedBy,
		instance.SessionAPIToken,
	)
	if err != nil {
		return 0, errors.Wrap(err, "compute create")
	}

	computeState, err := r.cloudSupplier.ComputeGet(ctx, instance.CloudName())
	if err != nil {
		return 0, errors.Wrap(err, "compute get")
	}

	if computeState.GetStatus() != compute.Instance_RUNNING || len(computeState.GetNetworkInterfaces()) == 0 {
		return time.Second * 3, nil
	}

	instance.CloudInstanceID = cloudInstanceID

	address := net.ParseIP(
		computeState.GetNetworkInterfaces()[0].GetPrimaryV4Address().GetOneToOneNat().GetAddress(),
	)
	if address == nil {
		return time.Second * 3, nil
	}

	instance.IP = address.String()
	instance.Status = instancerepo.InstanceStatusStarted

	return 0, r.saveAndReportInstance(
		ctx,
		instance,
		"VM created, but neko <b>is not</b> ready yet, wait ~4 minutes",
		false,
	)
}

func (r *Service) waitForNekoStart(ctx context.Context, instance *instancerepo.Instance) (time.Duration, error) {
	_, err := r.nekosupplier.GetStats(ctx, instance.IP, instance.SessionAPIToken)
	if err != nil {
		zerolog.Ctx(ctx).
			Warn().
			Err(err).
			Msg("neko.get.stats.err")

		return time.Second * 10, nil
	}

	instance.Status = instancerepo.InstanceStatusRunning

	return 0, r.saveAndReportInstance(ctx, instance, "Neko ready!!!", true)
}

func requireDeletion(ctx context.Context, stats *nekosupplier.Stats) bool {
	notUsedFor := time.Now().UTC().Sub(stats.LastUsageAt())

	if stats.TotalUsers != 0 || stats.TotalAdmins != 0 {
		zerolog.Ctx(ctx).Info().
			Interface("stats", stats).
			Msg("neko.instance.using")

		return false
	}

	const maxIdle = time.Minute * 10

	if notUsedFor > maxIdle {
		return true
	}

	zerolog.Ctx(ctx).Info().
		Interface("stats", stats).
		Str("not_used_for", time_utils.RoundDuration(notUsedFor)).
		Msg("neko.instance.no.users")

	return false
}

func requireRegularReport(instance *instancerepo.Instance) bool {
	now := time.Now().UTC()
	return now.Sub(instance.CreatedAt) > 15*time.Minute && now.Sub(instance.UpdatedAt) > 45*time.Minute
}

func (r *Service) processRunning(ctx context.Context, instance *instancerepo.Instance) (time.Duration, error) {
	r.proxy.SetTarget(&url.URL{Scheme: "http", Host: instance.IP})

	stats, err := r.nekosupplier.GetStats(ctx, instance.IP, instance.SessionAPIToken)
	if err != nil {
		return 0, errors.Wrap(err, "neko get stats")
	}

	if !requireDeletion(ctx, &stats) {
		if requireRegularReport(instance) {
			instance.UpdatedAt = time.Now().UTC()

			err = r.saveAndReportInstance(ctx, instance, "", true)
			if err != nil {
				return 0, errors.Wrap(err, "report instance")
			}
		}

		return 20 * time.Second, nil
	}

	instance.Status = instancerepo.InstanceStatusDeleting

	zerolog.Ctx(ctx).
		Info().
		Msg("neko.instance.deleting")

	return 0, r.saveAndReportInstance(ctx, instance, "Deleting instance because of no usage", true)
}

func (r *Service) Delete(ctx context.Context, instanceID string) error {
	instance, err := r.instanceRepo.GetInstance(ctx, instanceID)
	if err != nil {
		return errors.Wrap(err, "get instance")
	}

	instance.Status = instancerepo.InstanceStatusDeleting

	zerolog.Ctx(ctx).
		Info().
		Msg("neko.instance.deleting")

	return r.saveAndReportInstance(ctx, instance, "Deleting instance by request", true)
}

func (r *Service) processDeleting(ctx context.Context, instance *instancerepo.Instance) (time.Duration, error) {
	err := r.cloudSupplier.ComputeDeleteWaited(ctx, instance.CloudInstanceID)
	if err != nil {
		return 0, errors.Wrap(err, "compute delete")
	}

	instance.Status = instancerepo.InstanceStatusDeleted

	return 0, r.saveAndReportInstance(ctx, instance, "Instance deleted", false)
}
