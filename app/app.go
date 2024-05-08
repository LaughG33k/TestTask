package app

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/LaughG33k/TestTask/iternal"
	"github.com/LaughG33k/TestTask/iternal/handler"
	"github.com/LaughG33k/TestTask/iternal/repository"
	"github.com/LaughG33k/TestTask/pkg/client"
	"github.com/LaughG33k/TestTask/pkg/client/psql"
	"github.com/LaughG33k/TestTask/pkg/loging"
	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattes/migrate/source/file"
)

func Run() {

	err := godotenv.Load()

	if err != nil {
		log.Panic(err)
	}

	logrus, err := loging.InitLogrus(os.Getenv("LOG_PATH"))

	if err != nil {
		log.Panic(err)
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

	migrations, err := migrate.New(
		"file://migrations/psql/carInfo",
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", os.Getenv("PSQL_USERNAME"), os.Getenv("PSQL_PASSWORD"), os.Getenv("PSQL_HOST"), os.Getenv("PSQL_PORT"), os.Getenv("PSQL_DB_NAME"), "disable"),
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
		TLSConfig:    nil,
		PoolMaxConns: 1000,
	}, 15*time.Second)

	if err != nil {
		logrus.Panic(err)
		return
	}

	defer psql.Close()

	carRepo := repository.NewCarRepository(psql)

	client.ApiUrl = os.Getenv("CAR_INFO_API_URL")

	handl := handler.NewCarHandler(logrus, route, carRepo)

	handl.OperationTimeout = serverCfg.WriteTimeout

	handl.Start()

	if err := srv.StartServer(); err != nil {
		logrus.Panic(err)
	}
}
