package managerservice

import (
	"context"
	"neko-manager/pkg/instancerepo"
	"neko-manager/pkg/nekosupplier"

	"github.com/pkg/errors"
	tele "gopkg.in/telebot.v4"
)

func (r *Service) saveAndReportInstance(
	ctx context.Context,
	instance *instancerepo.Instance,
	text string,
	withStats bool) error {
	err := r.instanceRepo.SaveInstance(ctx, instance)
	if err != nil {
		return errors.Wrap(err, "save instance")
	}

	err = r.reportInstance(ctx, instance, text, withStats)
	if err != nil {
		return errors.Wrap(err, "report instance")
	}

	return nil
}

func (r *Service) reportInstance(
	ctx context.Context,
	instance *instancerepo.Instance,
	text string,
	withStats bool,
) error {
	var statsPtr *nekosupplier.Stats
	if withStats && instance.IP != "" {
		stats, err := r.nekosupplier.GetStats(ctx, instance.IP, instance.SessionAPIToken)
		if err == nil {
			statsPtr = &stats
		}
	}

	msgText, err := instance.Repr(statsPtr)
	if err != nil {
		return errors.Wrap(err, "instance repr")
	}

	if text != "" {
		msgText = text + "\n\n" + msgText
	}

	_, err = r.bot.Send(
		tele.ChatID(instance.TGChatID),
		msgText,
		&tele.SendOptions{ThreadID: instance.TGThreadChatID, ParseMode: tele.ModeHTML},
	)
	if err != nil {
		return errors.Wrap(err, "send tg message")
	}

	return nil
}
