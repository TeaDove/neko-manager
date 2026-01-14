package tgbotpresentation

import (
	"context"
	"fmt"
	"neko-manager/pkg/instancerepo"
	"neko-manager/pkg/managerservice"
	"neko-manager/pkg/nekosupplier"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/telebot_utils"
	tele "gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/middleware"
)

type Presentation struct {
	managerService *managerservice.Service
	nekosupplier   *nekosupplier.Supplier

	bot *tele.Bot

	allowedChats []int64
}

func New(
	managerService *managerservice.Service,
	bot *tele.Bot,
	nekosupplier *nekosupplier.Supplier,
	allowedChats []int64,
) *Presentation {
	return &Presentation{
		managerService: managerService,
		bot:            bot,
		nekosupplier:   nekosupplier,
		allowedChats:   allowedChats,
	}
}

func BuildBot(token string) (*tele.Bot, error) {
	bot, err := tele.NewBot(tele.Settings{
		Token:     token,
		ParseMode: tele.ModeHTML,
		OnError:   telebot_utils.ReportOnErr,
	})
	if err != nil {
		return nil, errors.Wrap(err, "new terx bot")
	}

	return bot, nil
}

func (r *Presentation) Run(ctx context.Context) {
	handlers := r.bot.Group()
	handlers.Use(middleware.Recover())
	handlers.Use(middleware.Whitelist(r.allowedChats...))

	handlers.Handle("/start", func(c tele.Context) error {
		return c.Reply(fmt.Sprintf(
			`Help:
/request [%s]- creates neko instance
/list - lists active instances
/delete &lt;id&gt; - deletes instance"`,
			strings.Join(instancerepo.ResourcesSizeNames(), ", ")))
	})

	handlers.Handle("/request", r.cmdRequest)
	handlers.Handle("/list", r.cmdList)
	handlers.Handle("/delete", r.cmdDelete)

	zerolog.Ctx(ctx).
		Info().
		Str("bot", r.bot.Me.Username).
		Msg("bot.polling.starting")
	r.bot.Start()
}

func (r *Presentation) cmdRequest(c tele.Context) error {
	var (
		resourceSize = instancerepo.ResourcesSizeM
		err          error
	)
	if len(c.Args()) != 0 {
		resourceSize, err = instancerepo.ParseResourcesSize(c.Args()[0])
		if err != nil {
			return telebot_utils.NewClientError(
				errors.New(
					"wrong resource size, allowed are: " + strings.Join(instancerepo.ResourcesSizeNames(), ", "),
				),
			)
		}
	}

	var threadId *int

	if c.ThreadID() != 0 {
		threadIdV := c.ThreadID()
		threadId = &threadIdV
	}

	_, err = r.managerService.RequestInstance(
		telebot_utils.GetOrSetCtx(c),
		c.Chat().ID,
		threadId,
		c.Sender().Username,
		resourceSize,
	)
	if err != nil {
		return errors.Wrap(err, "request instance")
	}

	return c.Reply("Instance requested, wait ~5 minutes")
}

func (r *Presentation) cmdList(c tele.Context) error {
	ctx := telebot_utils.GetOrSetCtx(c)

	instances, err := r.managerService.ListInstances(ctx)
	if err != nil {
		return errors.Wrap(err, "list instance")
	}

	if len(instances) == 0 {
		return c.Reply("No active instances")
	}

	for _, instance := range instances {
		tgreport, err := r.managerService.MakeTGReport(ctx, &instance, "", true)
		if err != nil {
			return errors.Wrap(err, "make tgreport")
		}

		err = c.Reply(tgreport)
		if err != nil {
			return errors.Wrap(err, "reply")
		}
	}

	return nil
}

func (r *Presentation) cmdDelete(c tele.Context) error {
	if len(c.Args()) != 1 {
		return telebot_utils.NewClientError(errors.New("Usage: /delete <id>"))
	}

	instanceID := c.Args()[0]

	err := r.managerService.Delete(telebot_utils.GetOrSetCtx(c), instanceID)
	if err != nil {
		return errors.Wrap(err, "delete instance")
	}

	return c.Reply("Instance deleting")
}
