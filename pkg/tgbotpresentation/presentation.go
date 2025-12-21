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
}

func New(managerService *managerservice.Service, terx *terx.Terx, nekosupplier *nekosupplier.Supplier) *Presentation {
	return &Presentation{managerService: managerService, terx: terx, nekosupplier: nekosupplier}
}

func (r *Presentation) Run() {
	r.terx.AddHandler(terx.FilterCommand("start"), func(c *terx.Ctx) error {
		return c.Reply("Help:\n/request - creates neko instance\n/list - lists active instances")
	})

	r.terx.AddHandler(terx.FilterAnd(terx.FilterCommand("request"), terx.FilterFromUser(r.terx.OwnerUserID)),
		func(c *terx.Ctx) error {
			_, err := r.managerService.RequestInstance(c.Context, c.Chat.ID, c.SentFrom.UserName)
			if err != nil {
				return errors.Wrap(err, "request instance")
			}

			return c.Replyf("Instance requested, wait ~5 minutes")
		})

	r.terx.AddHandler(terx.FilterCommand("list"),
		func(c *terx.Ctx) error {
			instances, err := r.managerService.ListInstances(c.Context)
			if err != nil {
				return errors.Wrap(err, "list instance")
			}

			for _, instance := range instances {
				var statsPtr *nekosupplier.Stats

				if instance.IP != "" {
					stats, err := r.nekosupplier.GetStats(c.Context, instance.IP, instance.SessionAPIToken)
					if err == nil {
						statsPtr = &stats
					}
				}

				err = c.Reply(instance.Repr(statsPtr))
				if err != nil {
					return errors.Wrap(err, "reply")
				}
			}

			return nil
		})

	r.terx.PollerRun()
}
