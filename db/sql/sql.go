package sql

import (
	"context"
	"database/sql"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	strcase "github.com/stoewer/go-strcase"

	"github.com/tochka/core-kit/apikit"
	"github.com/tochka/core-kit/errors"
	"github.com/tochka/core-kit/metrics"
	"github.com/tochka/core-kit/metrics/provider"
)

var (
	initMetricsOnce = sync.Once{}
	metricsLatency  metrics.Histogram
	metricsErrors   metrics.Counter
)

// initMetrics it's lazy initialize metrics it's used for overide DefaultProvider
func initMetrics() {
	initMetricsOnce.Do(func() {
		metricsLatency = provider.DefaultProvider.NewHistogram("sql_latency", "dsn", "query")
		metricsErrors = provider.DefaultProvider.NewCounter("sql_error", "dsn")
	})
}

// Open is the same as sql.Open, but returns an *sqlx.DB instead.
func Open(driverName, dsn string, user string, password string) (*DB, error) {
	str := strings.Replace(dsn, "{user}", user, -1)
	str = strings.Replace(str, "{password}", password, -1)

	db, err := sqlx.Open(driverName, str)
	if err != nil {
		return nil, errors.Wrap(err, "component", "sql_connection", "method", "open_db_connection")
	}
	db.MapperFunc(strcase.SnakeCase)

	initMetrics()
	return &DB{
		dsn:    dsn,
		sqlxDB: db,
	}, nil
}

// DB is a wrapper around sql.DB which keeps track of the driverName upon Open,
// used mostly to automatically bind named queries using the right bindvars.
type DB struct {
	dsn    string
	sqlxDB *sqlx.DB
}

// Select using this DB.
// Any placeholder parameters are replaced with supplied args.
func (db *DB) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) (err error) {
	t := metrics.NewTimer(metricsLatency.With("dsn", db.dsn, "query", query))
	defer t.ObserveDuration()
	errCtx := errors.Context("component", "sql_connection", "method", "select", "query", query, "dsn", db.dsn)
	defer errors.Defer(&err, errCtx)

	err = db.sqlxDB.SelectContext(ctx, dest, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.Wrap(apikit.ErrEntityNotFound)
		}
		metricsErrors.With("dsn", db.dsn).Add(1)
		return errors.Wrap(err)
	}
	return nil
}

// Get using this DB.
// Any placeholder parameters are replaced with supplied args.
// An error is returned if the result set is empty.
func (db *DB) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) (err error) {
	t := metrics.NewTimer(metricsLatency.With("dsn", db.dsn, "query", query))
	defer t.ObserveDuration()
	errCtx := errors.Context("component", "sql_connection", "method", "get", "query", query, "dsn", db.dsn)
	defer errors.Defer(&err, errCtx)

	err = db.sqlxDB.GetContext(ctx, dest, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return apikit.ErrEntityNotFound
		}
		metricsErrors.With("dsn", db.dsn).Add(1)
		return err
	}
	return nil
}

func (db *DB) In(ctx context.Context, dest interface{}, query string, args ...interface{}) (err error) {
	t := metrics.NewTimer(metricsLatency.With("dsn", db.dsn, "query", query))
	defer t.ObserveDuration()
	errCtx := errors.Context("component", "sql_connection", "method", "in", "query", query, "dsn", db.dsn)
	defer errors.Defer(&err, errCtx)

	query, args, err = sqlx.In(query, args...)
	if err != nil {
		errCtx.Add("action", "sqlx.in")
		return err
	}

	query = db.sqlxDB.Rebind(query)
	err = db.sqlxDB.SelectContext(ctx, dest, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return apikit.ErrEntityNotFound
		}
		metricsErrors.With("dsn", db.dsn).Add(1)
		return err
	}
	return nil
}

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (db *DB) Exec(ctx context.Context, query string, args ...interface{}) (result sql.Result, err error) {
	t := metrics.NewTimer(metricsLatency.With("dsn", db.dsn, "query", query))
	defer t.ObserveDuration()
	errCtx := errors.Context("component", "sql_connection", "method", "exec", "query", query, "dsn", db.dsn)
	defer errors.Defer(&err, errCtx)

	result, err = db.sqlxDB.ExecContext(ctx, query, args...)
	if err != nil {
		metricsErrors.With("dsn", db.dsn).Add(1)
		return nil, err
	}
	return result, nil
}

func (db *DB) Begin() (_ *Tx, err error) {
	t := metrics.NewTimer(metricsLatency.With("dsn", db.dsn, "query", "begin tx"))
	defer t.ObserveDuration()

	errCtx := errors.Context("component", "sql_connection", "method", "begin", "dsn", db.dsn)
	defer errors.Defer(&err, errCtx)

	tx, err := db.sqlxDB.Beginx()
	if err != nil {
		metricsErrors.With("dsn", db.dsn).Add(1)
		return nil, err
	}
	return &Tx{tx: tx, dsn: db.dsn}, err
}

// DriverName returns the driverName passed to the Open function for this DB.
func (db *DB) DriverName() string {
	return db.sqlxDB.DriverName()
}

// DB returns the sql.DB instance.
func (db *DB) DB() *sql.DB {
	return db.sqlxDB.DB
}

// Ping verifies a connection to the database is still alive,
// establishing a connection if necessary.
func (db *DB) Ping(ctx context.Context) error {
	err := db.sqlxDB.PingContext(ctx)
	if err != nil {
		return errors.Wrap(err, "component", "sql_connection", "method", "ping", "dsn", db.dsn)
	}
	return nil
}

// Tx is an sqlx wrapper around sql.Tx with extra functionality
type Tx struct {
	tx  *sqlx.Tx
	dsn string
}

func (tx *Tx) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) (err error) {
	t := metrics.NewTimer(metricsLatency.With("dsn", tx.dsn, "query", query))
	defer t.ObserveDuration()
	errCtx := errors.Context("component", "sql_connection", "method", "get", "query", query, "dsn", tx.dsn)
	defer errors.Defer(&err, errCtx)

	err = tx.tx.GetContext(ctx, dest, query, args)
	if err != nil {
		if err == sql.ErrNoRows {
			return apikit.ErrEntityNotFound
		}
		metricsErrors.With("dsn", tx.dsn).Add(1)
		return err
	}
	return nil
}

func (tx *Tx) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) (err error) {
	t := metrics.NewTimer(metricsLatency.With("dsn", tx.dsn, "query", query))
	defer t.ObserveDuration()
	errCtx := errors.Context("component", "sql_connection", "method", "select", "query", query, "dsn", tx.dsn)
	defer errors.Defer(&err, errCtx)

	err = tx.tx.SelectContext(ctx, dest, query, args)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.Wrap(apikit.ErrEntityNotFound)
		}
		metricsErrors.With("dsn", tx.dsn).Add(1)
		return errors.Wrap(err)
	}
	return nil
}

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (tx *Tx) Exec(ctx context.Context, query string, args ...interface{}) (result sql.Result, err error) {
	t := metrics.NewTimer(metricsLatency.With("dsn", tx.dsn, "query", query))
	defer t.ObserveDuration()
	errCtx := errors.Context("component", "sql_connection", "method", "exec", "query", query, "dsn", tx.dsn)
	defer errors.Defer(&err, errCtx)

	result, err = tx.tx.ExecContext(ctx, query, args...)
	if err != nil {
		metricsErrors.With("dsn", tx.dsn).Add(1)
		return nil, err
	}
	return result, nil
}

func (tx *Tx) Commit() (err error) {
	t := metrics.NewTimer(metricsLatency.With("dsn", tx.dsn, "query", "commit"))
	defer t.ObserveDuration()

	errCtx := errors.Context("component", "sql_connection", "method", "commit", "dsn", tx.dsn)
	defer errors.Defer(&err, errCtx)

	if err = tx.tx.Commit(); err != nil {
		metricsErrors.With("dsn", tx.dsn).Add(1)
		return err
	}
	return nil
}

func (tx *Tx) Rollback() (err error) {
	t := metrics.NewTimer(metricsLatency.With("dsn", tx.dsn, "query", "rollback"))
	defer t.ObserveDuration()

	errCtx := errors.Context("component", "sql_connection", "method", "rollback", "dsn", tx.dsn)
	defer errors.Defer(&err, errCtx)

	if err = tx.tx.Rollback(); err != nil {
		metricsErrors.With("dsn", tx.dsn).Add(1)
		return err
	}
	return nil
}
