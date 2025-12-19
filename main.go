package main

import (
	"neko-manager/pkg/instancerepo"
	"neko-manager/pkg/managerservice"

	"github.com/pkg/errors"
	"github.com/teadove/teasutils/service_utils/logger_utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Container struct {
	ManagerService *managerservice.Service
}

func build() (*Container, error) {
	db, err := gorm.Open(sqlite.Open(".data/db.sqlite"), &gorm.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "open gorm db")
	}

	err = db.AutoMigrate(new(instancerepo.Instance))
	if err != nil {
		return nil, errors.Wrap(err, "auto migrate")
	}

	instanceRepo := instancerepo.New(db)
	managerService := managerservice.New(instanceRepo)
	return &Container{ManagerService: managerService}, nil
}

func main() {
	ctx := logger_utils.NewLoggedCtx()

	container, err := build()
	if err != nil {
		panic(errors.Wrap(err, "build"))
	}

	err = container.ManagerService.RequestInstance(ctx)
	if err != nil {
		panic(errors.Wrap(err, "run"))
	}
}
