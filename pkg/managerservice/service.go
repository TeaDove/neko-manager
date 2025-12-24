package managerservice

import (
	"context"
	"neko-manager/pkg/cloudsupplier"
	"neko-manager/pkg/instancerepo"
	"neko-manager/pkg/nekoproxy"
	"neko-manager/pkg/nekosupplier"

	tele "gopkg.in/telebot.v4"
)

type Service struct {
	instanceRepo  *instancerepo.Repo
	cloudSupplier *cloudsupplier.Supplier
	nekosupplier  *nekosupplier.Supplier
	proxy         *nekoproxy.Proxy

	bot *tele.Bot
}

func New(
	instanceRepo *instancerepo.Repo,
	cloudSupplier *cloudsupplier.Supplier,
	bot *tele.Bot,
	nekosupplier *nekosupplier.Supplier,
	proxy *nekoproxy.Proxy,
) *Service {
	return &Service{
		instanceRepo:  instanceRepo,
		cloudSupplier: cloudSupplier,
		bot:           bot,
		nekosupplier:  nekosupplier,
		proxy:         proxy,
	}
}

func (r *Service) ListInstances(ctx context.Context) ([]instancerepo.Instance, error) {
	return r.instanceRepo.ListActiveInstances(ctx)
}
