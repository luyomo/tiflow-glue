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

package txn

import (
	"context"
	"net/url"
	"sync/atomic"
	"fmt"
	"os"
	"os/exec"

	"github.com/pingcap/errors"
	"github.com/pingcap/tiflow/cdc/model"
	"github.com/pingcap/tiflow/cdc/sinkv2/eventsink"
	"github.com/pingcap/tiflow/cdc/sinkv2/eventsink/txn/mysql"
	"github.com/pingcap/tiflow/cdc/sinkv2/eventsink/txn/postgres"
	"github.com/pingcap/tiflow/cdc/sinkv2/metrics"
	"github.com/pingcap/tiflow/cdc/sinkv2/tablesink/state"
	"github.com/pingcap/tiflow/pkg/causality"
	"github.com/pingcap/tiflow/pkg/config"
	psink "github.com/pingcap/tiflow/pkg/sink"
	pmysql "github.com/pingcap/tiflow/pkg/sink/mysql"

	"github.com/pingcap/log"
	"go.uber.org/zap"

        "github.com/hashicorp/go-plugin"
        "github.com/pingcap/tiflow/cdc/sinkv2/eventsink/txn/plugins/shared"
        "github.com/pingcap/tiflow/cdc/sinkv2/eventsink/txn/plugins/proto"
)

const (
	// DefaultConflictDetectorSlots indicates the default slot count of conflict detector.
	DefaultConflictDetectorSlots uint64 = 16 * 1024
)

// Assert EventSink[E event.TableEvent] implementation
var _ eventsink.EventSink[*model.SingleTableTxn] = (*sink)(nil)

// sink is the sink for SingleTableTxn.
type sink struct {
	conflictDetector *causality.ConflictDetector[*worker, *txnEvent]
	workers          []*worker
	cancel           func()
	// set when the sink is closed explicitly. and then subsequence `WriteEvents` call
	// should return an error.
	closed int32

	statistics *metrics.Statistics
}

func newSink(ctx context.Context, backends []backend, errCh chan<- error, conflictDetectorSlots uint64) *sink {
	workers := make([]*worker, 0, len(backends))
	for i, backend := range backends {
		w := newWorker(ctx, i, backend, errCh, len(backends))
		w.runBackgroundLoop()
		workers = append(workers, w)
	}
	detector := causality.NewConflictDetector[*worker, *txnEvent](workers, conflictDetectorSlots)
	return &sink{conflictDetector: detector, workers: workers}
}

// NewMySQLSink creates a mysql sink with given parameters.
func NewMySQLSink(
	ctx context.Context,
	sinkURI *url.URL,
	replicaConfig *config.ReplicaConfig,
	errCh chan<- error,
	conflictDetectorSlots uint64,
) (*sink, error) {
	var getConn pmysql.Factory = pmysql.CreateMySQLDBConn

	ctx1, cancel := context.WithCancel(ctx)
	statistics := metrics.NewStatistics(ctx1, psink.TxnSink)
	backendImpls, err := mysql.NewMySQLBackends(ctx, sinkURI, replicaConfig, getConn, statistics)
	if err != nil {
		cancel()
		return nil, err
	}

	backends := make([]backend, 0, len(backendImpls))
	for _, impl := range backendImpls {
		backends = append(backends, impl)
	}
	sink := newSink(ctx, backends, errCh, conflictDetectorSlots)
	sink.statistics = statistics
	sink.cancel = cancel

	return sink, nil
}

// NewMySQLSink creates a postgres sink with given parameters.
func NewPostgresSink(
	ctx context.Context,
	sinkURI *url.URL,
	replicaConfig *config.ReplicaConfig,
	errCh chan<- error,
	conflictDetectorSlots uint64,
) (*sink, error) {
	var getConn pmysql.Factory = pmysql.CreateMySQLDBConn

	ctx1, cancel := context.WithCancel(ctx)
	statistics := metrics.NewStatistics(ctx1, psink.TxnSink)
	backendImpls, err := postgres.NewPostgresBackends(ctx, sinkURI, replicaConfig, getConn, statistics)
	if err != nil {
		cancel()
		return nil, err
	}

	backends := make([]backend, 0, len(backendImpls))
	for _, impl := range backendImpls {
		backends = append(backends, impl)
	}
	sink := newSink(ctx, backends, errCh, conflictDetectorSlots)
	sink.statistics = statistics
	sink.cancel = cancel

	return sink, nil
}

// WriteEvents writes events to the sink.
func (s *sink) WriteEvents(txnEvents ...*eventsink.TxnCallbackableEvent) error {
	log.Info("DEBUG: WriteEvents", zap.Stack("tracestack"))
	if atomic.LoadInt32(&s.closed) != 0 {
		return errors.Trace(errors.New("closed sink"))
	}

	for _, txn := range txnEvents {
		log.Info("DEBUG: txn event", zap.String("event", fmt.Sprintf("%#v", txn)))
		if txn.GetTableSinkState() != state.TableSinkSinking {
			// The table where the event comes from is in stopping, so it's safe
			// to drop the event directly.
			txn.Callback()
			continue
		}
		log.Info("DEBUG: txn Event", zap.String("txn", fmt.Sprintf("%#v", txn.Event)))
		log.Info("DEBUG: txn Event Table", zap.String("txn", fmt.Sprintf("%#v", txn.Event.Table)))
		log.Info("DEBUG: txn Event Table Schema", zap.String("txn", fmt.Sprintf("%#v", txn.Event.Table.Schema)))
		log.Info("DEBUG: txn Event Table Table", zap.String("txn", fmt.Sprintf("%#v", txn.Event.Table.Table)))
		log.Info("DEBUG: txn Event Table TableID", zap.String("txn", fmt.Sprintf("%#v", txn.Event.Table.TableID)))
		log.Info("DEBUG: txn Event Table IsPartition", zap.String("txn", fmt.Sprintf("%#v", txn.Event.Table.IsPartition)))

		log.Info("DEBUG: txn Event TableInfo", zap.String("txn", fmt.Sprintf("%#v", txn.Event.TableInfo)))
		log.Info("DEBUG: txn Event StartTs", zap.String("txn", fmt.Sprintf("%#v", txn.Event.StartTs)))
		log.Info("DEBUG: txn Event CommitTs", zap.String("txn", fmt.Sprintf("%#v", txn.Event.CommitTs)))
		for _, row := range txn.Event.Rows {
		    // txn="&model.RowChangedEvent{StartTs:0x61992de57ac0002, CommitTs:0x61992de57ac0003, RowID:113, Table:(*model.TableName)(0xc0047aa300), ColInfos:[]rowcodec.ColInfo{rowcodec.ColInfo{ID:1, IsPKHandle:true, VirtualGenCol:false, Ft:(*types.FieldType)(0xc002d2c690)}}, TableInfo:(*model.TableInfo)(0xc002befad0), Columns:[]*model.Column{(*model.Column)(0xc0044f5380)}, PreColumns:[]*model.Column(nil), IndexColumns:[][]int{[]int{0}}, ApproximateDataSize:25, SplitTxn:false, ReplicatingTs:0x61992ddf9e80004}"
		    log.Info("DEBUG: txn Event Rows", zap.String("txn", fmt.Sprintf("%#v", row)))
		    log.Info("DEBUG: row change StartTs", zap.String("txn", fmt.Sprintf("%#v", row.StartTs)))
		    log.Info("DEBUG: row change CommitTs", zap.String("txn", fmt.Sprintf("%#v", row.CommitTs)))
		    log.Info("DEBUG: row change RowID", zap.String("txn", fmt.Sprintf("%#v", row.RowID)))
		    log.Info("DEBUG: row change Table", zap.String("txn", fmt.Sprintf("%#v", row.Table)))
		    log.Info("DEBUG: row change ColInfos", zap.String("txn", fmt.Sprintf("%#v", row.ColInfos)))
		    log.Info("DEBUG: row change TableInfo", zap.String("txn", fmt.Sprintf("%#v", row.TableInfo)))
		    for _, column := range row.Columns {
		        log.Info("DEBUG: row change Column", zap.String("txn", fmt.Sprintf("%#v", column)))
	            }
		    log.Info("DEBUG: row change PreColumns", zap.String("txn", fmt.Sprintf("%#v", row.PreColumns)))
		    log.Info("DEBUG: row change IndexColumns", zap.String("txn", fmt.Sprintf("%#v", row.IndexColumns )))
		    //  [txn="&model.Column{Name:\"col01\", Type:0x3, Charset:\"binary\", Flag:0xb, Value:116, Default:interface {}(nil), ApproximateBytes:160}"]
		}
		s.sendToPlugin(txn)

		s.conflictDetector.Add(newTxnEvent(txn))
	}
	return nil
}

// Close closes the sink. It won't wait for all pending items backend handled.
func (s *sink) Close() error {
	atomic.StoreInt32(&s.closed, 1)
	s.conflictDetector.Close()
	for _, w := range s.workers {
		w.Close()
	}
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	if s.statistics != nil {
		s.statistics.Close()
	}
	return nil
}

func (s *sink) sendToPlugin(txn *eventsink.TxnCallbackableEvent) {
    log.Info("DEBUG: txn event", zap.String("txn", fmt.Sprintf("%#v", *txn)))
    log.Info("DEBUG: callback event", zap.String("txn", fmt.Sprintf("%#v", *txn.Event)))

    // We're a host. Start by launching the plugin process.
    client := plugin.NewClient(&plugin.ClientConfig{
            HandshakeConfig: shared.Handshake,
            Plugins:         shared.PluginMap,
            Cmd:             exec.Command("sh", "-c", os.Getenv("KV_PLUGIN")),
            AllowedProtocols: []plugin.Protocol{
                    plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
    })
    defer client.Kill()

    // Connect via RPC
    rpcClient, err := client.Client()
    if err != nil {
        log.Info("DEBUG: Error:", zap.String("client", err.Error()))
        os.Exit(1)
    }

    // Request the plugin
    raw, err := rpcClient.Dispense("kv_grpc")
    if err != nil {
        log.Info("DEBUG: Error:", zap.String("raw", err.Error()))
        os.Exit(1)
    }
    kv := raw.(shared.KV)
		//log.Info("DEBUG: txn Event", zap.String("txn", fmt.Sprintf("%#v", txn.Event)))
		//log.Info("DEBUG: txn Event Table", zap.String("txn", fmt.Sprintf("%#v", txn.Event.Table)))
		//log.Info("DEBUG: txn Event Table Schema", zap.String("txn", fmt.Sprintf("%#v", txn.Event.Table.Schema)))
		//log.Info("DEBUG: txn Event Table Table", zap.String("txn", fmt.Sprintf("%#v", txn.Event.Table.Table)))
		//log.Info("DEBUG: txn Event Table TableID", zap.String("txn", fmt.Sprintf("%#v", txn.Event.Table.TableID)))
		//log.Info("DEBUG: txn Event Table IsPartition", zap.String("txn", fmt.Sprintf("%#v", txn.Event.Table.IsPartition)))

		//log.Info("DEBUG: txn Event TableInfo", zap.String("txn", fmt.Sprintf("%#v", txn.Event.TableInfo)))
		//log.Info("DEBUG: txn Event StartTs", zap.String("txn", fmt.Sprintf("%#v", txn.Event.StartTs)))
		//log.Info("DEBUG: txn Event CommitTs", zap.String("txn", fmt.Sprintf("%#v", txn.Event.CommitTs)))
		//log.Info("DEBUG: txn Event Rows", zap.String("txn", fmt.Sprintf("%#v", txn.Event.Rows)))

    // ColInfos
    var rowChangeEvents []*proto.RowChangeEvent
    for _, row := range txn.Event.Rows {
    // txn="&model.RowChangedEvent{StartTs:0x61992de57ac0002, CommitTs:0x61992de57ac0003, RowID:113, Table:(*model.TableName)(0xc0047aa300), ColInfos:[]rowcodec.ColInfo{rowcodec.ColInfo{ID:1, IsPKHandle:true, VirtualGenCol:false, Ft:(*types.FieldType)(0xc002d2c690)}}, TableInfo:(*model.TableInfo)(0xc002befad0), Columns:[]*model.Column{(*model.Column)(0xc0044f5380)}, PreColumns:[]*model.Column(nil), IndexColumns:[][]int{[]int{0}}, ApproximateDataSize:25, SplitTxn:false, ReplicatingTs:0x61992ddf9e80004}"
        log.Info("DEBUG: txn Event Rows", zap.String("txn", fmt.Sprintf("%#v", row)))
        log.Info("DEBUG: row chane StartTs", zap.String("txn", fmt.Sprintf("%#v", row.StartTs)))
        log.Info("DEBUG: row change CommitTs", zap.String("txn", fmt.Sprintf("%#v", row.CommitTs)))
        log.Info("DEBUG: row change RowID", zap.String("txn", fmt.Sprintf("%#v", row.RowID)))
        log.Info("DEBUG: row change Table", zap.String("txn", fmt.Sprintf("%#v", row.Table)))
        log.Info("DEBUG: row change ColInfos", zap.String("txn", fmt.Sprintf("%#v", row.ColInfos)))
	var colInfos []*proto.ColInfo
	for _, colInfo := range row.ColInfos {
            colInfos = append(colInfos, &proto.ColInfo {
                ID:            colInfo.ID,
		IsPKHandle:    colInfo.IsPKHandle,
		VirtualGenCol: colInfo.VirtualGenCol,
	    })
	}

        rowChangeEvents = append(rowChangeEvents, &proto.RowChangeEvent{
            StartTs: row.StartTs,
            CommitTs: row.CommitTs,
            RowID: row.RowID,
            ColInfos : colInfos,
        })
        log.Info("DEBUG: row change TableInfo", zap.String("txn", fmt.Sprintf("%#v", row.TableInfo)))
        for _, column := range row.Columns {
            log.Info("DEBUG: row change Column", zap.String("txn", fmt.Sprintf("%#v", column)))
               }
        log.Info("DEBUG: row change PreColumns", zap.String("txn", fmt.Sprintf("%#v", row.PreColumns)))
        log.Info("DEBUG: row change IndexColumns", zap.String("txn", fmt.Sprintf("%#v", row.IndexColumns )))
        //  [txn="&model.Column{Name:\"col01\", Type:0x3, Charset:\"binary\", Flag:0xb, Value:116, Default:interface {}(nil), ApproximateBytes:160}"]
    }

    err = kv.Put(&proto.PutRequest{
                TableName:   &proto.TableName {
                    Schema: txn.Event.Table.Schema ,
                    Table:  txn.Event.Table.Table,
                    TableID: txn.Event.Table.TableID,
                    IsPartition: txn.Event.Table.IsPartition,
                },
                RowChangeEvent: rowChangeEvents,
          })
    if err != nil {
        log.Info("DEBUG: Error:", zap.String("Put", err.Error()))
        os.Exit(1)
    }
    log.Info("DEBUG: plugin calling")
}

