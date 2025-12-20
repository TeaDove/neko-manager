package managerservice

import (
	"neko-manager/pkg/cloudsupplier"
	"neko-manager/pkg/instancerepo"
	"neko-manager/pkg/nekosupplier"

	"github.com/teadove/terx/terx"
)

type Service struct {
	instanceRepo  *instancerepo.Repo
	cloudSupplier *cloudsupplier.Supplier
	nekosupplier  *nekosupplier.Supplier

	terx *terx.Terx
}

func New(
	instanceRepo *instancerepo.Repo,
	cloudSupplier *cloudsupplier.Supplier,
	terx *terx.Terx,
	nekosupplier *nekosupplier.Supplier,
) *Service {
	return &Service{instanceRepo: instanceRepo, cloudSupplier: cloudSupplier, terx: terx, nekosupplier: nekosupplier}
}
