package managerservice

import (
	"context"
	"neko-manager/pkg/instancerepo"
)

type Service struct {
	instanceRepo *instancerepo.Repo
}

func New(instanceRepo *instancerepo.Repo) *Service {
	return &Service{instanceRepo: instanceRepo}
}

func (r *Service) RequestInstance(ctx context.Context) error {
	return nil
}
