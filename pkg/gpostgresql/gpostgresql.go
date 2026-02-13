package gpostgresql

import (
	"context"

	"github.com/HuseyinAsik/Notifications/pkg/logging"
	"github.com/HuseyinAsik/Notifications/pkg/settings"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Pool struct {
	Read  *pgxpool.Pool
	Write *pgxpool.Pool
}

var pool = &Pool{}

func (s *Pool) Close() {
	s.Read.Close()
	s.Write.Close()
}

type QueryTracer struct {
	logger *logging.LogWrapper
}

func (s *QueryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	s.logger.Debug(ctx, "Execute Query", zap.String("query", data.SQL), zap.Any("args", data.Args))
	return ctx
}

func (s *QueryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
}

func Setup(config *settings.Database, logger *logging.LogWrapper) {
	pool.Read = createPool(logger, config.ReadUrl, "read")
	pool.Write = createPool(logger, config.WriteUrl, "write")
}

func GetPool() *Pool {
	return pool
}

func createPool(logger *logging.LogWrapper, dsn, connType string) *pgxpool.Pool {
	ctx := context.Background()
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		logger.Fatal(ctx, "Unable to parse connection string for "+connType, zap.Error(err))
	}
	cfg.MaxConns = 100
	cfg.MinConns = 5
	cfg.ConnConfig.Tracer = &QueryTracer{logger: logger}

	pgpool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		logger.Fatal(ctx, "Unable to connect to database for "+connType, zap.Error(err))
	}
	return pgpool
}
