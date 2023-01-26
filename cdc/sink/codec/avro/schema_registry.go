// Copyright 2020 PingCAP, Inc.
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

package avro

import (
	"context"

	"github.com/linkedin/goavro/v2"
	"github.com/pingcap/tiflow/pkg/security"
	cerror "github.com/pingcap/tiflow/pkg/errors"
)

type schemaManager interface {
	GetCachedOrRegister(context.Context, string, uint64, SchemaGenerator) (*goavro.Codec, []byte, error)
	ClearRegistry(context.Context, string) error
}

// NewAvroSchemaManager creates a new schemaManager and test connectivity to the schema registry
func NewAvroSchemaManager(
	ctx context.Context, credential *security.Credential, registryURL, schemaRegistryProvider string, subjectSuffix string,
) (schemaManager, error) {

	switch schemaRegistryProvider {
		case "confluent":
			return NewConfluentSchemaManager(ctx, credential, registryURL, subjectSuffix)
		case "glue":
			return NewGlueSchemaManager(ctx, credential, registryURL, subjectSuffix)
	}
	return nil, cerror.ErrAvroSchemaAPIError.GenWithStack("No matched schema registry instance")
}

// SchemaGenerator represents a function that returns an Avro schema in JSON.
// Used for lazy evaluation
type SchemaGenerator func() (string, error)

