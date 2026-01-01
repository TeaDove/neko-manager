package managerservice

import (
	"context"
	"neko-manager/pkg/cloudsupplier"
	"neko-manager/pkg/instancerepo"
	"neko-manager/pkg/nekoproxy"
	"neko-manager/pkg/nekosupplier"
	"time"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	tele "gopkg.in/telebot.v4"
)

type Service struct {
	instanceRepo  *instancerepo.Repo
	cloudSupplier *cloudsupplier.Supplier
	nekosupplier  *nekosupplier.Supplier
	proxy         *nekoproxy.Proxy
	bot           *tele.Bot

	sleepOnErrDuration   time.Duration
	restartOnErrDuration time.Duration
	sizeToSpec           map[instancerepo.ResourcesSize]*compute.ResourcesSpec
}

func New(
	instanceRepo *instancerepo.Repo,
	cloudSupplier *cloudsupplier.Supplier,
	bot *tele.Bot,
	nekosupplier *nekosupplier.Supplier,
	proxy *nekoproxy.Proxy,
) *Service {
	return &Service{
		instanceRepo:         instanceRepo,
		cloudSupplier:        cloudSupplier,
		bot:                  bot,
		nekosupplier:         nekosupplier,
		proxy:                proxy,
		sleepOnErrDuration:   5 * time.Second,
		restartOnErrDuration: 7 * time.Minute,
		sizeToSpec: map[instancerepo.ResourcesSize]*compute.ResourcesSpec{
			instancerepo.ResourcesSizeS: {
				Memory:       1024 * 1024 * 1024 * 4,
				Cores:        4,
				CoreFraction: 100,
			},
			instancerepo.ResourcesSizeM: {
				Memory:       1024 * 1024 * 1024 * 8,
				Cores:        8,
				CoreFraction: 100,
			},
			instancerepo.ResourcesSizeL: {
				Memory:       1024 * 1024 * 1024 * 16,
				Cores:        16,
				CoreFraction: 100,
			},
		},
	}
}

func (r *Service) ListInstances(ctx context.Context) ([]instancerepo.Instance, error) {
	return r.instanceRepo.ListActiveInstances(ctx)
}
