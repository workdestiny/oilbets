package app

import (
	"database/sql"
	"net/http"

	"github.com/workdestiny/amlporn/config"
	"github.com/workdestiny/amlporn/entity"

	"github.com/acoshift/pgsql"
	"github.com/workdestiny/amlporn/repository"
	"github.com/moonrhythm/hime"
)

func notificationGetHandler(ctx *hime.Context) error {
	p := page(ctx)
	return ctx.View("app/notification", p)
}

func ajaxListNotificationPostHandler(ctx *hime.Context) error {

	id := GetMyID(ctx)

	var req entity.RequestNotificationNextLoad
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	res := entity.ResponseNotification{}

	if req.Next.String() == "1970-01-01 00:00:00 +0000 UTC" {

		list, err := repository.ListNotification(db, id, config.LimitListNotification+1)
		must(err)

		if len(list) == 0 {
			return ctx.NoContent()
		}

		if len(list) > config.LimitListNotification {
			res.Next = list[config.LimitListNotification-1].CreatedAt
			res.IsNext = true
			list = list[:config.LimitListNotification]
		}

		count, err := repository.CountNotification(db, id)
		must(err)

		res.Notification = list
		res.Count = count

		return ctx.Status(http.StatusOK).JSON(&res)
	}

	list, err := repository.ListNotificationNextLoad(db, id, req.Next, config.LimitListNotification+1)
	must(err)

	if len(list) == 0 {
		return ctx.NoContent()
	}

	if len(list) > config.LimitListNotification {
		res.Next = list[config.LimitListNotification-1].CreatedAt
		res.IsNext = true
		list = list[:config.LimitListNotification]
	}

	res.Notification = list

	return ctx.Status(http.StatusOK).JSON(&res)
}

func ajaxListNotificationTypePostHandler(ctx *hime.Context) error {

	id := GetMyID(ctx)

	var req entity.RequestNotificationNextLoad
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	res := entity.ResponseNotification{}

	if req.Next.String() == "1970-01-01 00:00:00 +0000 UTC" {

		list, err := repository.ListNotificationType(db, id, req.Type, config.LimitListNotification+1)
		must(err)

		if len(list) == 0 {
			return ctx.NoContent()
		}

		if len(list) > config.LimitListNotification {
			res.Next = list[config.LimitListNotification-1].CreatedAt
			res.IsNext = true
			list = list[:config.LimitListNotification]
		}

		res.Notification = list

		return ctx.Status(http.StatusOK).JSON(&res)
	}

	list, err := repository.ListNotificationTypeNextLoad(db, id, req.Type, req.Next, config.LimitListNotification+1)
	must(err)

	if len(list) == 0 {
		return ctx.NoContent()
	}

	if len(list) > config.LimitListNotification {
		res.Next = list[config.LimitListNotification-1].CreatedAt
		res.IsNext = true
		list = list[:config.LimitListNotification]
	}

	res.Notification = list

	return ctx.Status(http.StatusOK).JSON(&res)
}

func ajaxReadAllNotificationPostHandler(ctx *hime.Context) error {

	id := GetMyID(ctx)

	err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err := repository.ReadAllNotification(tx, id)
		if err != nil {
			return err
		}

		return nil
	})
	must(err)

	return ctx.NoContent()
}

func ajaxReadNotificationPostHandler(ctx *hime.Context) error {

	id := GetMyID(ctx)

	var req entity.Request
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	err = repository.CheckNotification(db, req.ID)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err := repository.ReadNotification(tx, id, req.ID)
		if err != nil {
			return err
		}

		return nil
	})
	must(err)

	return ctx.NoContent()
}

func ajaxResetNotificationPostHandler(ctx *hime.Context) error {

	id := GetMyID(ctx)

	err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err := repository.ResetNotification(tx, id)
		if err != nil {
			return err
		}

		return nil
	})
	must(err)

	return ctx.NoContent()
}
