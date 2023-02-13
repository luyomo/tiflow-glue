// Copyright 2022 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package postgres

import (
	"context"
	"database/sql"
	"net/url"
	"time"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
        "github.com/jackc/pgx/v5"
        "github.com/jackc/pgx/v5/pgxpool"

	"github.com/pingcap/errors"
	"github.com/pingcap/failpoint"
	"github.com/pingcap/log"
	timodel "github.com/pingcap/tidb/parser/model"
	"github.com/pingcap/tiflow/cdc/contextutil"
	"github.com/pingcap/tiflow/cdc/model"
	"github.com/pingcap/tiflow/cdc/sinkv2/ddlsink"
	"github.com/pingcap/tiflow/cdc/sinkv2/metrics"
	"github.com/pingcap/tiflow/pkg/config"
	cerror "github.com/pingcap/tiflow/pkg/errors"
	"github.com/pingcap/tiflow/pkg/errorutil"
	"github.com/pingcap/tiflow/pkg/quotes"
	"github.com/pingcap/tiflow/pkg/retry"
	"github.com/pingcap/tiflow/pkg/sink"
	pmysql "github.com/pingcap/tiflow/pkg/sink/mysql"
	"go.uber.org/zap"

	"github.com/pingcap/tidb/types"
	"github.com/pingcap/tidb/parser/mysql"
)

const (
	defaultDDLMaxRetry uint64 = 20

	// networkDriftDuration is used to construct a context timeout for database operations.
	networkDriftDuration = 5 * time.Second
)

// Assert DDLEventSink implementation
var _ ddlsink.DDLEventSink = (*postgresDDLSink)(nil)

type postgresDDLSink struct {
	// id indicates which processor (changefeed) this sink belongs to.
	id model.ChangeFeedID
	// db is the database connection.
	db  *sql.DB
	cfg *pmysql.Config
	// statistics is the statistics of this sink.
	// We use it to record the DDL count.
	statistics *metrics.Statistics
}

// NewMySQLDDLSink creates a new postgresDDLSink.
func NewPostgresDDLSink(
	ctx context.Context,
	sinkURI *url.URL,
	replicaConfig *config.ReplicaConfig,
	dbConnFactory pmysql.Factory,
) (*postgresDDLSink, error) {
	log.Info("DEBUG NewMySQLDDLSink", zap.Stack("stacktrace"))
	changefeedID := contextutil.ChangefeedIDFromCtx(ctx)
	cfg := pmysql.NewConfig()
	err := cfg.Apply(ctx, changefeedID, sinkURI, replicaConfig)
	if err != nil {
		return nil, err
	}

	dsnStr, err := pmysql.GenerateDSN(ctx, sinkURI, cfg, dbConnFactory)
	if err != nil {
		return nil, err
	}

	db, err := dbConnFactory(ctx, dsnStr)
	if err != nil {
		return nil, err
	}

	m := &postgresDDLSink{
		id:         changefeedID,
		db:         db,
		cfg:        cfg,
		statistics: metrics.NewStatistics(ctx, sink.TxnSink),
	}

	log.Info("MySQL DDL sink is created",
		zap.String("namespace", m.id.Namespace),
		zap.String("changefeed", m.id.ID))
	return m, nil
}

func (m *postgresDDLSink) WriteDDLEvent(ctx context.Context, ddl *model.DDLEvent) error {
	log.Info("DEBUG WriteDDLEvent", zap.Stack("stacktrace"))
	log.Info("DEBUG DDLEvent", zap.String("ddl event", fmt.Sprintf("%#v", ddl)))
	err := m.execDDLWithMaxRetries(ctx, ddl)
	return errors.Trace(err)
}

func (m *postgresDDLSink) execDDLWithMaxRetries(ctx context.Context, ddl *model.DDLEvent) error {
	return retry.Do(ctx, func() error {
		err := m.statistics.RecordDDLExecution(func() error { return m.execDDL(ctx, ddl) })
		if err != nil {
			if errorutil.IsIgnorableMySQLDDLError(err) {
				// NOTE: don't change the log, some tests depend on it.
				log.Info("Execute DDL failed, but error can be ignored",
					zap.Uint64("startTs", ddl.StartTs), zap.String("ddl", ddl.Query),
					zap.String("namespace", m.id.Namespace),
					zap.String("changefeed", m.id.ID),
					zap.Error(err))
				// If the error is ignorable, we will direly ignore the error.
				return nil
			}
			log.Warn("Execute DDL with error, retry later",
				zap.Uint64("startTs", ddl.StartTs), zap.String("ddl", ddl.Query),
				zap.String("namespace", m.id.Namespace),
				zap.String("changefeed", m.id.ID),
				zap.Error(err))
			return err
		}
		return nil
	}, retry.WithBackoffBaseDelay(pmysql.BackoffBaseDelay.Milliseconds()),
		retry.WithBackoffMaxDelay(pmysql.BackoffMaxDelay.Milliseconds()),
		retry.WithMaxTries(defaultDDLMaxRetry),
		retry.WithIsRetryableErr(cerror.IsRetryableError))
}

func (m *postgresDDLSink) execDDL(pctx context.Context, ddl *model.DDLEvent) error {
        connStr := "postgresql://pguser:1234Abcd@172.82.11.39/test?sslmode=disable"

        log.Info("Inserting data into pg")
        ctx := context.Background()
        var db *pgxpool.Pool

        db, err := pgxpool.New(ctx, connStr)
        defer db.Close()
        if err != nil {
            panic(err)
        }
        strDDL := m.generateDDL(pctx, ddl)
	batch := &pgx.Batch{}
	batch.Queue(strDDL)

        pgtx, err := db.Begin(ctx)
	if err != nil {
		return err
	}

        br := pgtx.SendBatch(context.Background(), batch)
        var berr error
        var result pgconn.CommandTag
        for berr == nil {
            result, berr = br.Exec()
            log.Info("result", zap.String("postgres batch", result.String() )  )
	    if berr != nil {
                   log.Info("result err", zap.String("postgres error",berr.Error() )  )
		   //cancelFunc()
	    }
        }
	br.Close()
        pgtx.Commit(ctx)

	writeTimeout, _ := time.ParseDuration(m.cfg.WriteTimeout)
	writeTimeout += networkDriftDuration
	ctx, cancelFunc := context.WithTimeout(pctx, writeTimeout)
	defer cancelFunc()

	shouldSwitchDB := needSwitchDB(ddl)

	failpoint.Inject("MySQLSinkExecDDLDelay", func() {
		select {
		case <-ctx.Done():
			failpoint.Return(ctx.Err())
		case <-time.After(time.Hour):
		}
		failpoint.Return(nil)
	})

	start := time.Now()
	log.Info("Start exec DDL", zap.Any("DDL", ddl), zap.String("namespace", m.id.Namespace),
		zap.String("changefeed", m.id.ID))
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if shouldSwitchDB {
		_, err = tx.ExecContext(ctx, "USE "+quotes.QuoteName(ddl.TableInfo.TableName.Schema)+";")
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Error("Failed to rollback", zap.String("namespace", m.id.Namespace),
					zap.String("changefeed", m.id.ID), zap.Error(err))
			}
			return err
		}
	}

	if _, err = tx.ExecContext(ctx, ddl.Query); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Error("Failed to rollback", zap.String("sql", ddl.Query),
				zap.String("namespace", m.id.Namespace),
				zap.String("changefeed", m.id.ID), zap.Error(err))
		}
		return err
	}

	if err = tx.Commit(); err != nil {
		log.Error("Failed to exec DDL", zap.String("sql", ddl.Query),
			zap.Duration("duration", time.Since(start)),
			zap.String("namespace", m.id.Namespace),
			zap.String("changefeed", m.id.ID), zap.Error(err))
		return cerror.WrapError(cerror.ErrMySQLTxnError, err)
	}

	log.Info("Exec DDL succeeded", zap.String("sql", ddl.Query),
		zap.Duration("duration", time.Since(start)),
		zap.String("namespace", m.id.Namespace),
		zap.String("changefeed", m.id.ID))
	return nil
}

func (m *postgresDDLSink) generateDDL(pctx context.Context, ddl *model.DDLEvent) string {
    value01, value02, value03 := ddl.TableInfo.GetRowColInfos()
    log.Info("generateDDL", zap.String("value01", fmt.Sprintf("%#v", value01) ))
    log.Info("generateDDL", zap.String("value02", fmt.Sprintf("%#v", value02) ))
    for _, fieldType := range value02 {
        log.Info("generateDDL", zap.String("field type", fmt.Sprintf("%#v", fieldType) ))
    }
    log.Info("generateDDL", zap.String("value03", fmt.Sprintf("%#v", value03) ))
    var arrCol []string
    for _,  col := range value03 {
        log.Info("generateDDL", zap.String("col info", fmt.Sprintf("%#v", col) ))
        log.Info("generateDDL", zap.String("field type", fmt.Sprintf("%#v", *col.Ft) ))
	colInfo, _ := ddl.TableInfo.GetColumnInfo(col.ID)
        log.Info("generateDDL", zap.String("col detail info", fmt.Sprintf("%#v", colInfo) ))
	strCol := fmt.Sprintf("%s %s ", colInfo.Name.L, m.mapDataType(&colInfo.FieldType) )
	arrCol = append(arrCol, strCol)
    }

    strPK := m.generatePK(ddl)
    log.Info("The PK is: ", zap.String("pk", strPK))
    strCols := strings.Join(arrCol, ",")
    strGenDDL := fmt.Sprintf("CREATE TABLE %s.%s(%s %s)", ddl.TableInfo.TableName.Schema, ddl.TableInfo.TableName.Table, strCols, strPK)
    log.Info("postgres ddl: ", zap.String("ddl", strGenDDL))
    // GetColumnInfo
//    for _, col := range ddl.TableInfo.cols {
//        log.Info("columns: ", zap.String("column", fmt.Sprintf("%#v", col ) ))
//    }
    return strGenDDL
}

func (m *postgresDDLSink) generatePK( ddl *model.DDLEvent) string {
    _, _,  colsInfo := ddl.TableInfo.GetRowColInfos()
    var arrCol []string
    for _,  col := range colsInfo {
	colInfo, _ := ddl.TableInfo.GetColumnInfo(col.ID)
        log.Info("generateDDL", zap.String("col detail info", fmt.Sprintf("%#v", colInfo) ))
        if mysql.HasPriKeyFlag(colInfo.FieldType.GetFlag()) {
	    arrCol = append(arrCol, colInfo.Name.L)

        }
    }
    strCols := strings.Join(arrCol, ",")
    if len(strCols) == 0 {
        return ""
    } else {
        return fmt.Sprintf(", primary key (%s)", strCols)
    }
}


func (m *postgresDDLSink) mapDataType(fieldType *types.FieldType) string {
    strNotNull := ""
    if mysql.HasNotNullFlag(fieldType.GetFlag()) {
        strNotNull = "not null"
    }

    switch fieldType.GetType() {
        case mysql.TypeLong:
            return fmt.Sprintf("%s %s", " int ", strNotNull)
        default:
            return ""

//     TypeUnspecified byte = 0
//     TypeTiny        byte = 1 // TINYINT
//     TypeShort       byte = 2 // SMALLINT
//     TypeLong        byte = 3 // INT
//     TypeFloat       byte = 4
//     TypeDouble      byte = 5
//     TypeNull        byte = 6
//     TypeTimestamp   byte = 7
//     TypeLonglong    byte = 8 // BIGINT
//     TypeInt24       byte = 9 // MEDIUMINT
//     TypeDate        byte = 10
//     /* TypeDuration original name was TypeTime, renamed to TypeDuration to resolve the conflict with Go type Time.*/
//     TypeDuration byte = 11
//     TypeDatetime byte = 12
//     TypeYear     byte = 13
//     TypeNewDate  byte = 14
//     TypeVarchar  byte = 15
//     TypeBit      byte = 16
// 
//     TypeJSON       byte = 0xf5
//     TypeNewDecimal byte = 0xf6
//     TypeEnum       byte = 0xf7
//     TypeSet        byte = 0xf8
//     TypeTinyBlob   byte = 0xf9
//     TypeMediumBlob byte = 0xfa
//     TypeLongBlob   byte = 0xfb
//     TypeBlob       byte = 0xfc
//     TypeVarString  byte = 0xfd
//     TypeString     byte = 0xfe /* TypeString is char type */
//     TypeGeometry   byte = 0xff

    }
}

func needSwitchDB(ddl *model.DDLEvent) bool {
	if len(ddl.TableInfo.TableName.Schema) == 0 {
		return false
	}
	if ddl.Type == timodel.ActionCreateSchema || ddl.Type == timodel.ActionDropSchema {
		return false
	}
	return true
}

func (m *postgresDDLSink) WriteCheckpointTs(_ context.Context, _ uint64, _ []*model.TableInfo) error {
	// Only for RowSink for now.
	return nil
}

// Close closes the database connection.
func (m *postgresDDLSink) Close() error {
	if err := m.db.Close(); err != nil {
		return errors.Trace(err)
	}
	if m.statistics != nil {
		m.statistics.Close()
	}

	return nil
}
