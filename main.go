package main

import (
	"context"
	"neko-manager/pkg/cloudsupplier"
	"neko-manager/pkg/instancerepo"
	"neko-manager/pkg/managerservice"
	"neko-manager/pkg/settings"

	"github.com/pkg/errors"
	"github.com/teadove/teasutils/service_utils/db_utils"
	"github.com/teadove/teasutils/service_utils/logger_utils"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/go-sdk/v2/credentials"
	"github.com/yandex-cloud/go-sdk/v2/pkg/options"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Container struct {
	ManagerService *managerservice.Service
}

func build(ctx context.Context) (*Container, error) {
	db, err := gorm.Open(sqlite.Open(settings.Settings.DB),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{SingularTable: true},
			TranslateError: true,
			Logger:         db_utils.ZerologAdapter{},
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "open gorm db")
	}

	err = db.AutoMigrate(new(instancerepo.Instance))
	if err != nil {
		return nil, errors.Wrap(err, "auto migrate")
	}

	instanceRepo := instancerepo.New(db)

	sdk, err := ycsdk.Build(ctx,
		options.WithCredentials(credentials.OAuthToken(settings.Settings.YCOauthToken)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "ycsdk build")
	}

	cloudSupplier, err := cloudsupplier.New(ctx, sdk, settings.Settings.YCFolderID)
	if err != nil {
		return nil, errors.Wrap(err, "new cloudsupplier")
	}

	managerService := managerservice.New(instanceRepo, cloudSupplier)

	return &Container{ManagerService: managerService}, nil
}

func main() {
	ctx := logger_utils.NewLoggedCtx()

	container, err := build(ctx)
	if err != nil {
		panic(errors.Wrap(err, "build"))
	}

	err = container.ManagerService.RequestInstance(ctx)
	if err != nil {
		panic(errors.Wrap(err, "run"))
	}
}
