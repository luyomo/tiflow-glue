// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
        "fmt"
        "io/ioutil"

        "github.com/hashicorp/go-plugin"
	"github.com/pingcap/tiflow/cdc/sinkv2/eventsink/txn/plugins/shared"
	"github.com/pingcap/tiflow/cdc/sinkv2/eventsink/txn/plugins/proto"
)

// Here is a real implementation of KV that writes to a local file with
// the key name and the contents are the value of the key.
type KV struct{}

func (KV) Put(putRequest *proto.PutRequest) error {
	value := []byte(fmt.Sprintf("%s\n\nWritten from plugin-go-grpc", fmt.Sprintf("%#v, %#v", putRequest.TableName, *putRequest.RowChangeEvent[0])))
        return ioutil.WriteFile("plugin-binlog", value, 0644)
}

func (KV) Get(key string) ([]byte, error) {
        return ioutil.ReadFile("kv_" + key)
}

func main() {
        plugin.Serve(&plugin.ServeConfig{
                HandshakeConfig: shared.Handshake,
                Plugins: map[string]plugin.Plugin{
                        "kv": &shared.KVGRPCPlugin{Impl: &KV{}},
                },

                // A non-nil value here enables gRPC serving for this plugin...
                GRPCServer: plugin.DefaultGRPCServer,
        })
}
