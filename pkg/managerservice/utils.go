package managerservice

import (
	"bytes"
	"context"
	"neko-manager/pkg/instancerepo"
	"neko-manager/pkg/nekosupplier"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	tele "gopkg.in/telebot.v4"
)

func (r *Service) saveAndReportInstance(
	ctx context.Context,
	instance *instancerepo.Instance,
	text string,
	withStats bool) error {
	zerolog.Ctx(ctx).
		Info().
		Object("instance", instance).
		Str("text", text).
		Msg("instance.reporting")

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

func (r *Service) MakeTGReport(
	ctx context.Context,
	instance *instancerepo.Instance,
	text string,
	withStats bool,
) (any, error) {
	var (
		statsPtr   *nekosupplier.Stats
		screenshot []byte
	)

	if withStats && instance.IP != nil {
		stats, err := r.nekosupplier.GetStats(ctx, instance.ToSupplierDTO())
		if err == nil {
			statsPtr = &stats
		} else {
			zerolog.Ctx(ctx).Error().
				Stack().Err(err).
				Msg("failed.to.get.stats")
		}

		screenshot, err = r.nekosupplier.GetScreenshot(ctx, instance.ToSupplierDTO())
		if err != nil {
			zerolog.Ctx(ctx).Error().
				Stack().Err(err).
				Msg("failed.to.get.screenshot")
		}
	}

	msgText, err := instance.Repr(statsPtr)
	if err != nil {
		return nil, errors.Wrap(err, "instance repr")
	}

	if text != "" {
		msgText = text + "\n\n" + msgText
	}

	if len(screenshot) != 0 {
		return &tele.Photo{Caption: msgText, File: tele.FromReader(bytes.NewReader(screenshot))}, nil
	}

	return msgText, nil
}

func (r *Service) reportInstance(
	ctx context.Context,
	instance *instancerepo.Instance,
	text string,
	withStats bool,
) error {
	msg, err := r.MakeTGReport(ctx, instance, text, withStats)
	if err != nil {
		return errors.Wrap(err, "make tgreport")
	}

	opts := &tele.SendOptions{ParseMode: tele.ModeHTML}
	if instance.TGThreadChatID != nil {
		opts.ThreadID = *instance.TGThreadChatID
	}

	_, err = r.bot.Send(tele.ChatID(instance.TGChatID), msg, opts)
	if err != nil {
		return errors.Wrap(err, "send tg message")
	}

	return nil
}
