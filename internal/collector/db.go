package collector

import (
	"context"
	"database/sql"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/pkg/errors"
)

type DB struct {
	db     *sql.DB
	tables mapset.Set[string]
}

func NewDB(db *sql.DB) *DB {
	return &DB{db: db, tables: mapset.NewSet[string]()}
}

func (db *DB) Warmup(ctx context.Context) error {
	if _, err := db.db.Exec("SELECT 1"); err != nil {
		return errors.WithMessagef(err, "failed to warmup database")
	}

	rs, err := db.db.QueryContext(ctx, listAllExistingTables())
	if err != nil && err != sql.ErrNoRows {
		return errors.WithMessagef(err, "failed to list existing tables")
	}

	defer rs.Close()
	for rs.Next() {
		var tableName string
		if err := rs.Scan(&tableName); err != nil {
			return errors.WithMessagef(err, "failed to scan table name")
		}
		db.tables.Add(tableName)
	}

	return nil
}

func (db *DB) createTables(ctx context.Context, projectName, experimentName string) error {
  txn := ctx.Value(Txn).(*sql.Tx)

	tableName := projectName + "-" + experimentName + "_log"
	if db.tables.Contains(tableName) {
		return nil
	}
	_, err := txn.ExecContext(ctx, createTableQuery(projectName, experimentName))
	if err != nil {
		return errors.WithMessagef(err, "failed to create table")
	}
	_, err = txn.ExecContext(ctx, createIndexQuery(projectName, experimentName))
	if err != nil {
		return errors.WithMessagef(err, "failed to create index")
	}

	db.tables.Add(tableName)

	return nil
}

func (db *DB) InsertLog(ctx context.Context, ltc LogToCollect) error {
  txn := ctx.Value(Txn).(*sql.Tx)

	if err := db.createTables(ctx, ltc.ProjectName, ltc.ExperimentName); err != nil {
		return errors.WithMessagef(err, "failed to create table")
	}

	if _, err := txn.ExecContext(ctx,
		insertLogQuery(ltc.ProjectName, ltc.ExperimentName),
		ltc.Timestamp,
		ltc.Source,
		ltc.Line,
		ltc.ExperimentName,
		ltc.ProjectName,
		ltc.RunName,
		ltc.ContainerID,
		ltc.NodeRank,
	); err != nil {
		return errors.WithMessagef(err, "failed to insert log")
	}

	return nil
}

func (db *DB) InsertStop(ctx context.Context, soe StopOrErr) error {
  txn := ctx.Value(Txn).(*sql.Tx)

	if err := db.createTables(ctx, soe.ProjectName, soe.ExperimentName); err != nil {
		return errors.WithMessagef(err, "failed to create table")
	}

	if _, err := txn.ExecContext(ctx,
		insertStopQuery(soe.ProjectName, soe.ExperimentName),
		soe.ExperimentName,
		soe.ProjectName,
		soe.RunName,
		soe.ContainerID,
		soe.NodeRank,
		soe.StopOrErr,
	); err != nil {
		return errors.WithMessagef(err, "failed to insert stop")
	}

	return nil
}
