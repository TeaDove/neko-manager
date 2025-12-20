package settings

import (
	"github.com/teadove/teasutils/service_utils/settings_utils"
)

type baseSettings struct {
	DB string `env:"DB" envDefault:".data/db.sqlite"`

	// https://oauth.yandex.ru/verification_code
	YCOauthToken string `env:"YC_OAUTH_TOKEN" envDefault:"BAD_TOKEN"`
	YCFolderID   string `env:"YC_FOLDER_ID" envDefault:"BAD_FOLDER"`

	SSHUserName  string `env:"SSH_USER_NAME" envDefault:"neko"`
	SSHPublicKey string `env:"SSH_PUBLIC_KEY" envDefault:"BAD_PUBLIC_KEY"`

	BotToken   string `env:"BOT_TOKEN" envDefault:"BAD_TOKEN"`
	BotOwnerID int64  `env:"BOT_OWNER_ID" envDefault:"-1"`
}

// Settings
// nolint: gochecknoglobals // need it
var Settings = settings_utils.MustGetSetting[baseSettings]("NEKO_")
