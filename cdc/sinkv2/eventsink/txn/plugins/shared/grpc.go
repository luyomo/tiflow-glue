// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shared

import (
	"github.com/pingcap/tiflow/cdc/sinkv2/eventsink/txn/plugins/proto"
        "golang.org/x/net/context"
)

// GRPCClient is an implementation of KV that talks over RPC.
type GRPCClient struct{ client proto.KVClient }

func (m *GRPCClient) Put(putRequest *proto.PutRequest) error {
        _, err := m.client.Put(context.Background(), putRequest)
        return err
}


//Schema      string `protobuf:"bytes,1,opt,name=Schema,proto3" json:"Schema,omitempty"`
//        Table       string `protobuf:"bytes,2,opt,name=Table,proto3" json:"Table,omitempty"`
//        TableID     int64  `protobuf:"varint,3,opt,name=TableID,proto3" json:"TableID,omitempty"`
//        IsPartition b
//

func (m *GRPCClient) Get(key string) ([]byte, error) {
        resp, err := m.client.Get(context.Background(), &proto.GetRequest{
                Key: key,
        })
        if err != nil {
                return nil, err
        }

        return resp.Value, nil
}

// Here is the gRPC server that GRPCClient talks to.
type GRPCServer struct {
        proto.UnimplementedKVServer
        // This is the real implementation
        Impl KV
}

func (m *GRPCServer) Put(
        ctx context.Context,
        req *proto.PutRequest) (*proto.Empty, error) {
        return &proto.Empty{}, m.Impl.Put(req)
}

func (m *GRPCServer) Get(
        ctx context.Context,
        req *proto.GetRequest) (*proto.GetResponse, error) {
        v, err := m.Impl.Get(req.Key)
        return &proto.GetResponse{Value: v}, err
}
