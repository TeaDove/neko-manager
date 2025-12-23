package tgbotpresentation

import (
	"neko-manager/pkg/managerservice"
	"neko-manager/pkg/nekosupplier"

	"github.com/pkg/errors"
	"github.com/teadove/terx/terx"
)

type Presentation struct {
	managerService *managerservice.Service
	nekosupplier   *nekosupplier.Supplier

	terx *terx.Terx

	allowedChatID int64
}

func New(
	managerService *managerservice.Service,
	terx *terx.Terx,
	nekosupplier *nekosupplier.Supplier,
	allowedChatID int64,
) *Presentation {
	return &Presentation{
		managerService: managerService,
		terx:           terx,
		nekosupplier:   nekosupplier,
		allowedChatID:  allowedChatID,
	}
}

func (r *Presentation) Run() { //nolint: gocognit // Presentation
	r.terx.AddHandler(terx.FilterCommand("start"), func(c *terx.Ctx) error {
		return c.Reply(
			"Help:\n/request - creates neko instance\n/list - lists active instances\n/delete &lt;id&gt; - deletes instance",
		)
	})

	filters := terx.FilterOr(terx.FilterFromUser(r.terx.OwnerUserID), terx.FilterFromChat(r.allowedChatID))

	r.terx.AddHandler(terx.FilterAnd(filters, terx.FilterCommand("request")),
		func(c *terx.Ctx) error {
			_, err := r.managerService.RequestInstance(c.Context, c.Chat.ID, c.SentFrom.UserName)
			if err != nil {
				return errors.Wrap(err, "request instance")
			}

			return c.Replyf("Instance requested, wait ~5 minutes")
		})

	r.terx.AddHandler(terx.FilterAnd(filters, terx.FilterCommand("list")),
		func(c *terx.Ctx) error {
			instances, err := r.managerService.ListInstances(c.Context)
			if err != nil {
				return errors.Wrap(err, "list instance")
			}

			if len(instances) == 0 {
				return c.Reply("No active instances")
			}

			for _, instance := range instances {
				var statsPtr *nekosupplier.Stats

				if instance.IP != "" {
					stats, err := r.nekosupplier.GetStats(c.Context, instance.IP, instance.SessionAPIToken)
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
		})

	r.terx.AddHandler(terx.FilterAnd(filters, terx.FilterCommand("delete")),
		func(c *terx.Ctx) error {
			instanceID := c.Text

			err := r.managerService.Delete(c.Context, instanceID)
			if err != nil {
				return errors.Wrap(err, "delete instance")
			}

			return c.Reply("Instance deleting")
		})

	r.terx.PollerRun()
}
