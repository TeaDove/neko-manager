package tgbotpresentation

import (
	"neko-manager/pkg/managerservice"

	"github.com/pkg/errors"
	"github.com/teadove/terx/terx"
)

type Presentation struct {
	managerService *managerservice.Service

	terx *terx.Terx
}

func New(managerService *managerservice.Service, terx *terx.Terx) *Presentation {
	return &Presentation{managerService: managerService, terx: terx}
}

func (r *Presentation) Run() {
	r.terx.AddHandler(terx.FilterCommand("start"), func(c *terx.Ctx) error {
		return c.Reply("Help:\n/request - creates neko instance\n/list - lists active instances")
	})

	r.terx.AddHandler(terx.FilterAnd(terx.FilterCommand("request"), terx.FilterFromUser(r.terx.OwnerUserID)),
		func(c *terx.Ctx) error {
			instance, err := r.managerService.RequestInstance(c.Context, c.Chat.ID, c.SentFrom.UserName)
			if err != nil {
				return errors.Wrap(err, "request instance")
			}

			return c.Replyf("Instance requested, id=<code>%s</code>", instance.ID)
		})

	r.terx.PollerRun()
}
