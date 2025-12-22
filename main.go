package main

import (
	"context"
	"neko-manager/pkg/cloudsupplier"
	"neko-manager/pkg/instancerepo"
	"neko-manager/pkg/managerservice"
	"neko-manager/pkg/nekoproxy"
	"neko-manager/pkg/nekosupplier"
	"neko-manager/pkg/settings"
	"neko-manager/pkg/tgbotpresentation"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/teadove/teasutils/service_utils/db_utils"
	"github.com/teadove/teasutils/service_utils/logger_utils"
	"github.com/teadove/terx/terx"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/go-sdk/v2/credentials"
	"github.com/yandex-cloud/go-sdk/v2/pkg/options"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Container struct {
	ManagerService    *managerservice.Service
	TGBotPresentation *tgbotpresentation.Presentation
	NekoProxy         *nekoproxy.Proxy
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

	cloudSupplier, err := cloudsupplier.New(
		ctx,
		sdk,
		settings.Settings.YCFolderID,
		settings.Settings.SSHPublicKey,
		settings.Settings.SSHUserName,
	)
	if err != nil {
		return nil, errors.Wrap(err, "new cloudsupplier")
	}

	terxBot, err := terx.New(terx.Config{
		Token:          settings.Settings.BotToken,
		OwnerUserID:    settings.Settings.BotOwnerID,
		ReplyWithErr:   true,
		SendErrToOwner: true,
	})
	if err != nil {
		return nil, errors.Wrap(err, "new terx bot")
	}

	nekoSupplier := nekosupplier.New(&http.Client{Timeout: time.Second * 5})
	nekoProxy := nekoproxy.New()

	managerService := managerservice.New(instanceRepo, cloudSupplier, terxBot, nekoSupplier, nekoProxy)

	tgBotPresentation := tgbotpresentation.New(managerService, terxBot, nekoSupplier)

	return &Container{ManagerService: managerService, TGBotPresentation: tgBotPresentation, NekoProxy: nekoProxy}, nil
}

func main() {
	ctx := logger_utils.NewLoggedCtx()

	container, err := build(ctx)
	if err != nil {
		panic(errors.Wrap(err, "build"))
	}

	err = container.ManagerService.DeleteUnknown(ctx)
	if err != nil {
		panic(errors.Wrap(err, "delete unknown instance"))
	}

	err = container.ManagerService.Reconciliation(ctx)
	if err != nil {
		panic(errors.Wrap(err, "reconciliation"))
	}

	go func() {
		if settings.Settings.ProxyURL == "" {
			return
		}

		container.NekoProxy.URL = settings.Settings.ProxyURL

		err = http.ListenAndServeTLS( //nolint: gosec // don't care
			":8080",
			settings.Settings.CertFile,
			settings.Settings.KeyFile,
			container.NekoProxy.MakeSTDProxy(ctx),
		)
		if err != nil {
			panic(errors.Wrap(err, "serve"))
		}
	}()

	container.TGBotPresentation.Run()
}
