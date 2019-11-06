package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	stdlog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dmitrymomot/examplesrv/client"

	"github.com/dmitrymomot/examplesrv/jobs"
	"github.com/dmitrymomot/examplesrv/pb/examplesrv"
	"github.com/dmitrymomot/examplesrv/service"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

const serviceName = "examplesrv"

var (
	buildTag string

	appName = flag.String("app_name", "undefined", "global application name")

	httpPort    = flag.Int("http_port", 8888, "http port")
	grpcPort    = flag.Int("grpc_port", 9200, "grpc port")
	debug       = flag.Bool("debug", false, "enable debug mode")
	tlsRootCert = flag.String("tls_root_cert", "", "path to TLS root certificate")

	dbHost     = flag.String("db_host", "postgres", "postgresql database name")
	dbPort     = flag.Int("db_port", 5432, "postgresql database port")
	dbName     = flag.String("db_name", "", "postgresql database name")
	dbUser     = flag.String("db_user", "", "postgresql database user name")
	dbPassword = flag.String("db_password", "", "postgresql database password")
	dbPoolSize = flag.Int("db_pool_size", 10, "postgresql connections pool size")

	redisHost = flag.String("redis_host", "redis://redis:6379", "redis host, e.g.: redis://redis:6379")

	natsHost         = flag.String("nats_host", "nats://nats-cluster:4222", "NATS host to connect, e.g.: nats://nats-cluster:4222")
	natsQueueSubject = flag.String("nats_queue_subject", serviceName, "NATS queue subject")

	grpcServer *grpc.Server
	httpServer *http.Server

	log *zap.SugaredLogger

	queueSubscr *nats.Subscription
)

func main() {
	flag.Parse()

	// Set up logger
	var zlogger *zap.Logger
	var err error
	if *debug {
		zlogger, err = zap.NewDevelopment()
	} else {
		zlogger, err = zap.NewProduction()
	}
	if err != nil {
		stdlog.Fatal("can't init logger", err)
	}
	defer zlogger.Sync()
	log = zlogger.Sugar().With(zap.String("service_name", serviceName), zap.String("build", buildTag))
	log.Info("starting app")
	log.Debug("debug mode enabled")

	// Listen interrupt signal from OS
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	// Context with the cancellation function
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Error group with custom context
	eg, ctx := errgroup.WithContext(ctx)

	// TLS config for do services
	var tlsConfig *tls.Config
	if *tlsRootCert != "" {
		tlsConfig, err = loadTLSRootCert(*tlsRootCert)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Set up datapase connection
	db := pg.Connect(&pg.Options{
		Addr:            fmt.Sprintf("%s:%d", *dbHost, *dbPort),
		User:            *dbUser,
		Password:        *dbPassword,
		Database:        *dbName,
		ApplicationName: *appName,
		PoolSize:        *dbPoolSize,
		TLSConfig:       tlsConfig,
	})
	defer db.Close()

	if *debug {
		// Log each database query
		db.AddQueryHook(newDBLogger(log))
	}

	if err := db.CreateTable(&struct {
		tableName string    `pg:"test_tbl"`
		ID        uuid.UUID `pg:",pk"`
	}{}, &orm.CreateTableOptions{IfNotExists: true}); err != nil {
		log.Fatal(err)
	}

	// Setup redis connections pool
	redisPool := newRedisPool(*redisHost, tlsConfig)

	// Init worker pool
	pool := work.NewWorkerPool(jobs.WorkerPoolContext{}, 10, "{default}", redisPool)
	// Init worker pool context
	worker := jobs.NewWorker(pool)
	if err := worker.RegisterJobs(); err != nil {
		log.Fatal(err)
	}
	// Start processing jobs
	pool.Start()
	// Stop the pool
	defer pool.Stop()
	// Init enqueuer
	enqueuer := work.NewEnqueuer("{default}", redisPool)

	// Set up NATS connection
	natsOpts := setupNatsConnOptions(log, []nats.Option{nats.Name(serviceName)})
	nc, err := nats.Connect(*natsHost, natsOpts...)
	if err != nil {
		log.Fatalw(err.Error(), zap.String("host", *natsHost))
	}
	defer nc.Close()

	// Subscribe to {{serviceName}}.* channel in {{serviceName}} queue
	queueSubscr, err := nc.QueueSubscribe(*natsQueueSubject, serviceName, func(msg *nats.Msg) {
		log.Infow("received message",
			zap.String("subject", msg.Subject),
			zap.String("reply", msg.Reply),
			zap.String("data", string(msg.Data)))
	})
	if err != nil {
		log.Fatal(err)
	}
	defer queueSubscr.Unsubscribe()

	eg.Go(func() error {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()
		uuid := uuid.New().String()
		for {
			select {
			case <-ticker.C:
				msg := fmt.Sprintf("publisher %s: Current time is %s", uuid, time.Now().String())
				nc.Publish(*natsQueueSubject, []byte(msg))
			case <-ctx.Done():
				return nil
			}
		}
	})

	// Init RPC service
	exmplsrv := service.New(log, db, enqueuer)

	// Set up grpc server
	grpcServer = grpc.NewServer()
	// Registering of grpc services
	examplesrv.RegisterServiceServer(grpcServer, exmplsrv)
	// Run gRPC server
	eg.Go(func() error {
		addr := fmt.Sprintf(":%d", *grpcPort)
		grpcLis, err := net.Listen("tcp", addr)
		if err != nil {
			return errors.Wrap(err, "grpc failed to listen")
		}
		log.Infof("gRPC server serving at %s", addr)
		if err := grpcServer.Serve(grpcLis); err != nil {
			return errors.Wrap(err, "grpc server")
		}
		return nil
	})

	// Set up a connection to the server.
	// c, err := e.NewDefaultClient("examplesrv:9200")
	// if err != nil {
	// 	log.Fatalf("did not connect: %v", err)
	// }
	c := client.Must(client.NewDefaultClient("examplesrv:9200"))
	defer c.Close()

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", *httpPort),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Registering of twirp rpc services
	exmplHandler := examplesrv.NewServiceServer(exmplsrv, nil)
	router.Mount(exmplHandler.PathPrefix(), exmplHandler)

	// Run http server
	eg.Go(func() error {
		log.Infof("HTTP server serving at %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			return errors.Wrap(err, "http server")
		}
		return nil
	})

	// Wait for interrupt signal or context cancellation
	select {
	case <-interrupt:
		break
	case <-ctx.Done():
		break
	}

	log.Info("received shutdown signal")

	if grpcServer != nil {
		grpcServer.GracefulStop()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if httpServer != nil {
		_ = httpServer.Shutdown(shutdownCtx)
	}

	if err := eg.Wait(); err != nil {
		log.Fatalf("failed to wait goroutine group: %v.", err.Error())
	}

	log.Infof("shutdown at %s", time.Now().String())
}

func loadTLSRootCert(path string) (*tls.Config, error) {
	caPool := x509.NewCertPool()
	severCert, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	caPool.AppendCertsFromPEM(severCert)
	return &tls.Config{
		RootCAs:            caPool,
		InsecureSkipVerify: true,
	}, nil
}

func newRedisPool(connStr string, tlsCnf *tls.Config) *redis.Pool {
	return &redis.Pool{
		MaxActive: 10,
		MaxIdle:   10,
		Wait:      true,
		Dial:      setupRedisConnection(connStr, tlsCnf),
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

func setupRedisConnection(connStr string, tlsCnf *tls.Config) func() (redis.Conn, error) {
	return func() (redis.Conn, error) {
		conn, err := redis.DialURL(connStr, redis.DialTLSConfig(tlsCnf))
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
}

func setupNatsConnOptions(log *zap.SugaredLogger, opts []nats.Option) []nats.Option {
	totalWait := 10 * time.Minute
	reconnectDelay := time.Second

	opts = append(opts, nats.ReconnectWait(reconnectDelay))
	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))
	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		log.Infof("Disconnected due to: %s, will attempt reconnects for %.0fm", err, totalWait.Minutes())
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Infof("Reconnected [%s]", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		log.Fatalf("Exiting: %v", nc.LastError())
	}))
	return opts
}
