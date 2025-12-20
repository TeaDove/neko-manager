package instancerepo

import (
	"context"

	"github.com/pkg/errors"

	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) SaveInstance(ctx context.Context, instance *Instance) error {
	err := r.db.WithContext(ctx).Save(instance).Error
	if err != nil {
		return errors.Wrap(err, "save instance")
	}

	return nil
}

func (r *Repo) ListActiveInstances(ctx context.Context) ([]Instance, error) {
	instances, err := gorm.G[Instance](r.db).Where("status != ?", InstanceStatusDeleted).Find(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "list active instances")
	}

	return instances, nil
}
