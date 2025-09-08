package database

import (
	"context"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

func InitDB() *pgxpool.Pool {
	cfg, _ := pgxpool.ParseConfig("")
	cfg.ConnConfig.Host = os.Getenv("HOST_DB")
	portInt, err := strconv.ParseInt(os.Getenv("PORT_DB"), 0, 16)
	cfg.ConnConfig.Port = uint16(portInt)
	cfg.ConnConfig.User = os.Getenv("POSTGRES_USER")
	cfg.ConnConfig.Password = os.Getenv("POSTGRES_PASSWORD")
	cfg.ConnConfig.Database = os.Getenv("POSTGRES_DB")
	cfg.MaxConns = 20

	dbPool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		logrus.WithError(err).Error("DB error")
		os.Exit(1)
	}

	return dbPool
}
