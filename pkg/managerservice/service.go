package managerservice

import (
	"context"
	"neko-manager/pkg/cloudsupplier"
	"neko-manager/pkg/instancerepo"
	"neko-manager/pkg/nekoproxy"
	"neko-manager/pkg/nekosupplier"

	"github.com/teadove/terx/terx"
)

type Service struct {
	instanceRepo  *instancerepo.Repo
	cloudSupplier *cloudsupplier.Supplier
	nekosupplier  *nekosupplier.Supplier
	proxy         *nekoproxy.Proxy

	terx *terx.Terx
}

func New(
	instanceRepo *instancerepo.Repo,
	cloudSupplier *cloudsupplier.Supplier,
	terx *terx.Terx,
	nekosupplier *nekosupplier.Supplier,
	proxy *nekoproxy.Proxy,
) *Service {
	return &Service{
		instanceRepo:  instanceRepo,
		cloudSupplier: cloudSupplier,
		terx:          terx,
		nekosupplier:  nekosupplier,
		proxy:         proxy,
	}
}

func (r *Service) ListInstances(ctx context.Context) ([]instancerepo.Instance, error) {
	return r.instanceRepo.ListActiveInstances(ctx)
}
