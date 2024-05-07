package app

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/testTask/iternal"
	"github.com/testTask/iternal/handler"
	"github.com/testTask/iternal/repository"
	"github.com/testTask/pkg/client/psql"
	"github.com/testTask/pkg/loging"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/mattes/migrate/source/file"
)

func Run() {

	logrus, err := loging.InitLogrus("/Users/user/Desktop/testTask/logs.txt")

	if err != nil {
		log.Panic(err)
	}

	err = godotenv.Load()

	if err != nil {
		logrus.Panic(err)
	}

	route := gin.Default()

	rt, err := strconv.Atoi(os.Getenv("READ_TIMEOUT_MILISSECONDS"))

	if err != nil {
		logrus.Panic(err)
	}

	wt, err := strconv.Atoi(os.Getenv("WRITE_TIMEOUT_MILLISECONDS"))

	if err != nil {
		logrus.Panic(err)
	}

	serverCfg := iternal.HttpServerConfig{
		Host:         os.Getenv("HOST"),
		Port:         os.Getenv("PORT"),
		ReadTimeout:  time.Duration(rt) * time.Millisecond,
		WriteTimeout: time.Duration(wt) * time.Millisecond,

		MaxHandlers: 1000,
		Hanler:      route,
	}

	srv, err := iternal.HttpServerInit(serverCfg)

	if err != nil {
		logrus.Panic(err)
	}

	defer srv.Close()

	port, err := strconv.Atoi(os.Getenv("PSQL_PORT"))

	if err != nil {
		logrus.Panic(err)
	}

	poolMaxConns, err := strconv.Atoi(os.Getenv("PSQL_POOL_MAX_CONNS"))

	if err != nil {
		logrus.Panic(err)
	}

	migrations, err := migrate.New(
		"file://migrations/psql/carInfo",
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", os.Getenv("PSQL_USERNAME"), os.Getenv("PSQL_PASSWORD"), os.Getenv("PSQL_HOST"), os.Getenv("PSQL_PORT"), os.Getenv("PSQL_DB_NAME"), os.Getenv("PSQL_SSLMODE")),
	)

	if err != nil {
		logrus.Panic()
	}

	defer migrations.Close()

	if err := migrations.Up(); err != nil {
		if err.Error() != "no change" {
			logrus.Panic(err)
		}
	}

	psql, err := psql.Newclient(psql.PsqlConnParams{
		Host:         os.Getenv("PSQL_HOST"),
		Port:         uint16(port),
		User:         os.Getenv("PSQL_USERNAME"),
		Password:     os.Getenv("PSQL_PASSWORD"),
		Db:           os.Getenv("PSQL_DB_NAME"),
		SslMode:      os.Getenv("PSQL_SSLMODE"),
		TLSConfig:    nil,
		PoolMaxConns: poolMaxConns,
	}, 15*time.Second)

	if err != nil {
		logrus.Panic(err)
		return
	}

	defer psql.Close()

	carRepo := repository.NewCarRepository(psql)

	handl := handler.NewCarHandler(logrus, route, carRepo)

	handl.OperationTimeout = serverCfg.WriteTimeout

	handl.Start()

	if err := srv.StartServer(); err != nil {
		logrus.Panic(err)
	}
}
