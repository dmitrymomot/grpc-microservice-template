package service

import (
	"github.com/dmitrymomot/examplesrv/pb/examplesrv"
	"github.com/go-pg/pg/v9"
	"github.com/gocraft/work"
	"go.uber.org/zap"
)

type service struct {
	examplesrv.UnimplementedServiceServer
	log *zap.SugaredLogger // uber's structured and leveled logger
	db  *pg.DB             // postgresql database connection
	wp  *work.Enqueuer     // redis worker pool
}

// New service factory
func New(log *zap.SugaredLogger, db *pg.DB, wp *work.Enqueuer) examplesrv.ServiceServer {
	return &service{log: log, db: db, wp: wp}
}
