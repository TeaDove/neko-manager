package tgbotpresentation

import (
	"context"
	"neko-manager/pkg/managerservice"
	"neko-manager/pkg/nekosupplier"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/service_utils/logger_utils"
	"github.com/teadove/teasutils/utils/redact_utils"
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

func GetOrSetCtx(c tele.Context) context.Context {
	ctx, ok := c.Get("ctx").(context.Context)
	if ok {
		return ctx
	}

	ctx = logger_utils.NewLoggedCtx()
	if c.Chat() != nil && c.Chat().Title != "" {
		ctx = logger_utils.WithValue(ctx, "in", c.Chat().Title)
	}

	if c.Text() != "" {
		ctx = logger_utils.WithValue(ctx, "text", redact_utils.Trim(c.Text()))
	}

	if c.Sender() != nil {
		ctx = logger_utils.WithValue(ctx, "from", c.Sender().Username)
	}

	c.Set("ctx", ctx)

	return ctx
}

func BuildBot(token string) (*tele.Bot, error) {
	bot, err := tele.NewBot(tele.Settings{
		Token:     token,
		ParseMode: tele.ModeHTML,
		OnError: func(err error, c tele.Context) {
			zerolog.Ctx(GetOrSetCtx(c)).
				Error().
				Stack().Err(err).
				Msg("failed.to.process.tg.update")
		},
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
		return c.Reply(
			"Help:\n/request - creates neko instance\n/list - lists active instances\n/delete &lt;id&gt; - deletes instance",
		)
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
	_, err := r.managerService.RequestInstance(GetOrSetCtx(c), c.Chat().ID, c.ThreadID(), c.Sender().Username)
	if err != nil {
		return errors.Wrap(err, "request instance")
	}

	return c.Reply("Instance requested, wait ~5 minutes")
}

func (r *Presentation) cmdList(c tele.Context) error {
	ctx := GetOrSetCtx(c)

	instances, err := r.managerService.ListInstances(ctx)
	if err != nil {
		return errors.Wrap(err, "list instance")
	}

	if len(instances) == 0 {
		return c.Reply("No active instances")
	}

	for _, instance := range instances {
		var statsPtr *nekosupplier.Stats

		if instance.IP != "" {
			stats, err := r.nekosupplier.GetStats(ctx, instance.IP, instance.SessionAPIToken)
			if err == nil {
				statsPtr = &stats
			}
		}

		text, err := instance.Repr(statsPtr)
		if err != nil {
			return errors.Wrap(err, "repr")
		}

		err = c.Reply(text)
		if err != nil {
			return errors.Wrap(err, "reply")
		}
	}

	return nil
}

func (r *Presentation) cmdDelete(c tele.Context) error {
	if len(c.Args()) != 1 {
		return c.Reply("Usage: /delete <id>")
	}

	instanceID := c.Args()[0]

	err := r.managerService.Delete(GetOrSetCtx(c), instanceID)
	if err != nil {
		return errors.Wrap(err, "delete instance")
	}

	return c.Reply("Instance deleting")
}
