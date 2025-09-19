package database

import (
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
)

func InitQueryBuilder() goqu.DialectWrapper {
	dialect := goqu.Dialect("postgres")

	return dialect
}
