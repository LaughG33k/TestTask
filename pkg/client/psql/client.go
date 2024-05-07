package psql

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/TestTask/pkg"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type PsqlConnParams struct {
	Host         string
	Port         uint16
	User         string
	Password     string
	Db           string
	SslMode      string
	TLSConfig    *tls.Config
	PoolMaxConns int
}

func Newclient(params PsqlConnParams, connTimeout time.Duration) (*pgxpool.Pool, error) {

	var conn *pgxpool.Pool

	ssl := "disable"

	if params.SslMode != "" {
		ssl = params.SslMode
	}

	cfg, err := pgxpool.ParseConfig(fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=%s pool_max_conns=%d", params.User, params.Password, params.Host, params.Port, params.Db, ssl, params.PoolMaxConns))

	if err != nil {
		return nil, err
	}

	cfg.HealthCheckPeriod = 10 * time.Second
	cfg.ConnConfig.TLSConfig = params.TLSConfig

	err = pkg.Retry(func() error {

		tm, canc := context.WithTimeout(context.Background(), connTimeout)

		defer canc()

		newConn, err := pgxpool.NewWithConfig(tm, cfg)
		if err != nil {
			return err
		}

		conn = newConn

		return nil

	}, 5, time.Second*3)

	if err != nil {
		return nil, err
	}

	return conn, nil

}
