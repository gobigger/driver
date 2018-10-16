package data_cockroach


import (
	. "github.com/yatlabs/bigger"
	"database/sql"
)

type (
	CockroachExecutor interface {
		Exec(query string, args ...Any) (sql.Result,error)
		Prepare(query string) (*sql.Stmt, error)
		Query(query string, args ...Any) (*sql.Rows, error)
		QueryRow(query string, args ...Any) (*sql.Row)
	}
)

