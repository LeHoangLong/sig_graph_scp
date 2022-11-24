package repository_server

import (
	"errors"
	"sig_graph_scp/pkg/utility"

	"github.com/jackc/pgconn"
	"go.uber.org/multierr"
	"gorm.io/gorm"
)

func wrapError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = multierr.Append(utility.ErrDatabase, utility.ErrNotFound)
	} else if err, ok := wrapPgError(err); ok {
		err = multierr.Append(utility.ErrDatabase, err)
	} else {
		err = multierr.Append(utility.ErrDatabase, err)
	}

	return err
}

type EPgErrorCode = string

const (
	EPgErrorCodeUniqueConstraintViolation EPgErrorCode = "23505"
)

func wrapPgError(err error) (error, bool) {
	pgErr := &pgconn.PgError{}
	if errors.As(err, &pgErr) {
		if pgErr.Code == EPgErrorCodeUniqueConstraintViolation {
			return utility.ErrAlreadyExists, true
		}
	}
	return err, false
}
