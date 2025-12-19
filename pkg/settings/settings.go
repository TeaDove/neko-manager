package settings

import (
	"github.com/teadove/teasutils/service_utils/settings_utils"
)

type baseSettings struct {
	DB string `env:"DB" envDefault:".data/db.sqlite"`

	// https://oauth.yandex.ru/verification_code
	YCOauthToken string `env:"YC_OAUTH_TOKEN" envDefault:"BAD_TOKEN"`
	YCFolderID   string `env:"YC_FOLDER_ID" envDefault:"BAD_FOLDER"`

	BotToken string `env:"BOT_TOKEN" envDefault:"BAD_TOKEN"`
}

// Settings
// nolint: gochecknoglobals // need it
var Settings = settings_utils.MustGetSetting[baseSettings]("NEKO_")
