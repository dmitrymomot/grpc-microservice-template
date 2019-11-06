package main

import (
	ctx "context"

	"github.com/go-pg/pg/v9"
	"go.uber.org/zap"
)

type dbLogger struct {
	log *zap.SugaredLogger
}

func newDBLogger(log *zap.SugaredLogger) dbLogger {
	return dbLogger{log: log}
}

func (d dbLogger) BeforeQuery(c ctx.Context, q *pg.QueryEvent) (ctx.Context, error) {
	return c, nil
}

func (d dbLogger) AfterQuery(c ctx.Context, q *pg.QueryEvent) error {
	qs, err := q.FormattedQuery()
	if err != nil {
		d.log.Errorf("query string error: %v", err.Error())
		return err
	}
	d.log.Debugf("query string: %s", qs)
	return nil
}
