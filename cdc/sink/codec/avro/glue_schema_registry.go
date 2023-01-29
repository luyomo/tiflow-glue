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
	"sync"
        "fmt"
        "runtime/debug"

	"github.com/linkedin/goavro/v2"
	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	cerror "github.com/pingcap/tiflow/pkg/errors"
	"github.com/pingcap/tiflow/pkg/security"
	"go.uber.org/zap"

        "github.com/aws/aws-sdk-go-v2/config"
        "github.com/aws/aws-sdk-go-v2/service/glue"
        "github.com/aws/aws-sdk-go-v2/service/glue/types"
        "github.com/aws/aws-sdk-go-v2/aws"

	"github.com/satori/go.uuid"
)

const VERSION = uint8(3)            // 3 is fixed for the glue message
const COMPRESSION_BYTE = uint8(0)    // 0  no compression


func NewGlueSchemaManager(ctx context.Context, credential *security.Credential, registryURL string, subjectSuffix string,
	)(schemaManager, error) {
	glueSchemaManager := GlueSchemaManager{
//		registryURL:   registryURL,
		cache:         make(map[string]*schemaCacheEntry, 1),
		subjectSuffix: subjectSuffix,
		registryName:  registryURL,
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
	    return nil, err
	}
	glueSchemaManager.glueClient = glue.NewFromConfig(cfg)

	if err := glueSchemaManager.setGlueRegistryID() ; err != nil {
		return nil, err
	}

	return &glueSchemaManager, nil
}

// ---------- Glue schema manager
// schemaManager is used to register Avro Schemas to the Registry server,
// look up local cache according to the table's name, and fetch from the Registry
// in cache the local cache entry is missing.
type GlueSchemaManager struct {
	registryURL   string
	subjectSuffix string

	credential *security.Credential // placeholder, currently always nil

	cacheRWLock sync.RWMutex
	cache       map[string]*schemaCacheEntry

	glueClient  *glue.Client

	registryName string
	registryArn *string
}

// Register a schema in schema registry, no cache
func (m *GlueSchemaManager) Register(
	ctx context.Context,
	topicName string,
	codec *goavro.Codec,
) ([]byte, error) {
	// The Schema Registry expects the JSON to be without newline characters

        debug.PrintStack()
        log.Info("DEBUG Info 05:",zap.String("stack",  string(debug.Stack()) ) )


    log.Info("Registering the schema into glue")
    listSchemas, err := m.glueClient.ListSchemas(context.TODO(), &glue.ListSchemasInput{RegistryId: &types.RegistryId{RegistryArn: m.registryArn}})
    if err != nil {
        return nil, err
    }
    for _, schema := range listSchemas.Schemas {
        if *schema.SchemaName == topicName {
            log.Info("Create new schema version: ", zap.String("schema arn", *schema.SchemaArn), zap.String("registry name", *schema.RegistryName), zap.String("schema name", topicName))
            registerSchemaVersion, err := m.glueClient.RegisterSchemaVersion(context.TODO(), &glue.RegisterSchemaVersionInput{SchemaDefinition: aws.String(codec.Schema()), SchemaId: &types.SchemaId{SchemaArn: schema.SchemaArn}})
            if err != nil {
                return nil, err
            }
	    // log.Info("Registered schema successfully",
	    // 	zap.Int("id", jsonResp.ID),
	    // 	zap.String("uri", uri),
	    // 	zap.ByteString("body", body))
	    uuidSchemaVersionId, err := uuid.FromString(*registerSchemaVersion.SchemaVersionId)
            if err != nil {
                return nil, err
            }
	    log.Info("Length of uuid ", zap.Int("uuid", len(uuidSchemaVersionId.Bytes())))

	var header []byte
	header = append(header, VERSION)
	header = append(header, COMPRESSION_BYTE)
	header = append(header, uuidSchemaVersionId.Bytes()...)


            // return *registerSchemaVersion.SchemaVersionId, nil
            return header, nil
        }
    }

    createSchema, err := m.glueClient.CreateSchema( context.TODO(), &glue.CreateSchemaInput{DataFormat: types.DataFormat("AVRO"), SchemaName: aws.String(topicName), RegistryId: &types.RegistryId{RegistryArn: m.registryArn}, SchemaDefinition: aws.String(codec.Schema()), Compatibility: types.CompatibilityBackward } )
    if err != nil {
        return nil, err
    }
    log.Info(fmt.Sprintf("Data: %#v ", createSchema))

	    uuidSchemaVersionId, err := uuid.FromString(*createSchema.SchemaVersionId)
            if err != nil {
                return nil, err
            }
	    log.Info("Length of uuid ", zap.Int("uuid", len(uuidSchemaVersionId.Bytes())))
	var header []byte
	header = append(header, VERSION)
	header = append(header, COMPRESSION_BYTE)
	header = append(header, uuidSchemaVersionId.Bytes()...)
            // return *registerSchemaVersion.SchemaVersionId, nil
            return header, nil
//	return *createSchema.SchemaVersionId, nil
}

func (m *GlueSchemaManager)Lookup( ctx context.Context, topicName string, tiSchemaID uint64) (*goavro.Codec, string, error){
	return nil, "", nil
}
// ClearRegistry clears the Registry subject for the given table. Should be idempotent.
// Exported for testing.
// NOT USED for now, reserved for future use.
func (m *GlueSchemaManager) ClearRegistry(ctx context.Context, topicName string) error {
	return nil
// 	log.Info("---------- ---------- ClearRegistry: topic name", zap.String("topic name", topicName))
// 	uri := m.registryURL + "/subjects/" + url.QueryEscape(
// 		m.topicNameToSchemaSubject(topicName),
// 	)
// 	req, err := http.NewRequestWithContext(ctx, "DELETE", uri, nil)
// 	if err != nil {
// 		log.Error("Could not construct request for clearRegistry", zap.String("uri", uri))
// 		return cerror.WrapError(cerror.ErrAvroSchemaAPIError, err)
// 	}
// 	req.Header.Add(
// 		"Accept",
// 		"application/vnd.schemaregistry.v1+json, application/vnd.schemaregistry+json, "+
// 			"application/json",
// 	)
// 	resp, err := httpRetry(ctx, m.credential, req)
// 	if err != nil {
// 		return err
// 	}
// 	defer func() {
// 		_, _ = io.Copy(io.Discard, resp.Body)
// 		_ = resp.Body.Close()
// 	}()
// 
// 	if resp.StatusCode == 200 {
// 		log.Info("Clearing Registry successful")
// 		return nil
// 	}
// 
// 	if resp.StatusCode == 404 {
// 		log.Info("Registry already cleaned")
// 		return nil
// 	}
// 
// 	log.Error("Error when clearing Registry", zap.Int("status", resp.StatusCode))
// 	return cerror.ErrAvroSchemaAPIError.GenWithStack(
// 		"Error when clearing Registry, status = %d",
// 		resp.StatusCode,
// 	)
}

// GetCachedOrRegister checks if the suitable Avro schema has been cached.
// If not, a new schema is generated, registered and cached.
// Re-registering an existing schema shall return the same id(and version), so even if the
// cache is out-of-sync with schema registry, we could reload it.
func (m *GlueSchemaManager) GetCachedOrRegister(
	ctx context.Context,
	topicName string,
	tiSchemaID uint64,
	schemaGen SchemaGenerator,
) (*goavro.Codec, []byte, error) {
	log.Info("---------- ---------- GetCachedOrRegister: ", zap.String("topic name", topicName), zap.Uint64("tiSchemaID", tiSchemaID))
        log.Info(fmt.Sprintf("The schema generator is <%#v>", schemaGen))
	key := m.topicNameToSchemaSubject(topicName)
	m.cacheRWLock.RLock()
        log.Info("Check schema: ", zap.String("subject", key), zap.Uint64("schema id", tiSchemaID))
	if entry, exists := m.cache[key]; exists && entry.tiSchemaID == tiSchemaID {
		log.Debug("Avro schema GetCachedOrRegister cache hit",
			zap.String("key", key),
			zap.Uint64("tiSchemaID", tiSchemaID),
			zap.String("registryID", entry.registryID))
		m.cacheRWLock.RUnlock()
		return entry.codec, []byte(string(entry.registryID)), nil
	}
	m.cacheRWLock.RUnlock()

	log.Info("Avro schema lookup cache miss",
		zap.String("key", key),
		zap.Uint64("tiSchemaID", tiSchemaID))

	schema, err := schemaGen()
	if err != nil {
		return nil, []byte(string("0")), err
	}

	codec, err := goavro.NewCodec(schema)
	if err != nil {
		log.Error("GetCachedOrRegister: Could not make goavro codec", zap.Error(err))
		return nil, []byte(string("0")), cerror.WrapError(cerror.ErrAvroSchemaAPIError, err)
	}

        log.Info(fmt.Sprintf("The code to be registered: %#v", codec))

	id, err := m.Register(ctx, key, codec)
	if err != nil {
		log.Error("GetCachedOrRegister: Could not register schema", zap.Error(err))
		return nil, []byte(string("0")), errors.Trace(err)
	}

	cacheEntry := new(schemaCacheEntry)
	cacheEntry.codec = codec
	cacheEntry.registryID = string(id)
	cacheEntry.tiSchemaID = tiSchemaID
        log.Info(fmt.Sprintf("cache key: %s, codec: %#v", key, codec))

	m.cacheRWLock.Lock()
	m.cache[key] = cacheEntry
	m.cacheRWLock.Unlock()

	log.Info("Avro schema GetCachedOrRegister successful with cache miss",
		zap.Uint64("tiSchemaID", cacheEntry.tiSchemaID),
		zap.String("registryID", cacheEntry.registryID),
		zap.String("schema", cacheEntry.codec.Schema()))

	return codec, id, nil
}

// TopicNameStrategy, ksqlDB only supports this
func (m *GlueSchemaManager) topicNameToSchemaSubject(topicName string) string {
	return topicName + m.subjectSuffix
}

func (m *GlueSchemaManager) setGlueRegistryID() (error) {
	listRegistries, err := m.glueClient.ListRegistries(context.TODO(), &glue.ListRegistriesInput{})
        if err != nil {
		return err
	}
	for _, registryItem := range listRegistries.Registries {
		if *registryItem.RegistryName == m.registryName {
			m.registryArn = registryItem.RegistryArn
			return nil
		}
	}
	return errors.New("No valid glue schema registry found.")
}
