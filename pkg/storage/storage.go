// Copyright © 2018 Alfred Chou <unioverlord@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	core "github.com/universonic/ivy-utils/pkg/storage/core"
	etcd "github.com/universonic/ivy-utils/pkg/storage/etcd"
	zap "go.uber.org/zap"
	yaml "gopkg.in/yaml.v2"
)

type StorageConfig struct {
	Adapter string                 `json:"adapter,omitempty" yaml:"adapter"`
	Config  map[string]interface{} `json:"config,omitempty" yaml:"config"`
}

func NewStorageConfig() *StorageConfig {
	return &StorageConfig{
		Config: make(map[string]interface{}),
	}
}

// NewStorageFromConfigBytes spawn storage with config from given json bytes
func NewStorageFromConfigBytes(config []byte, logger *zap.SugaredLogger) (core.Storage, error) {
	cfg := NewStorageConfig()
	err := json.Unmarshal(config, cfg)
	if err != nil {
		return nil, err
	}
	return NewStorage(cfg, logger)
}

// NewStorageFromConfigFile is a shorthand of NewStorage
func NewStorageFromConfigFile(fp string, logger *zap.SugaredLogger) (core.Storage, error) {
	fi, err := os.Open(fp)
	if err != nil {
		return nil, err
	}
	defer fi.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, fi)
	if err != nil {
		return nil, err
	}
	config := buf.Bytes()
	qualifiedConfig := NewStorageConfig()
	ext := filepath.Ext(fi.Name())
	switch ext {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(config, qualifiedConfig)
	case ".json":
		err = json.Unmarshal(config, qualifiedConfig)
	default:
		return nil, fmt.Errorf("Unrecognized configuration file extension: %s", ext)
	}
	if err != nil {
		return nil, err
	}
	return NewStorage(qualifiedConfig, logger)
}

// NewStorage is a helper which creates a database connection instance and returns
// any encountered error.
func NewStorage(config *StorageConfig, logger *zap.SugaredLogger) (core.Storage, error) {
	switch config.Adapter {
	case "etcd":
		etcd := etcd.New()
		b, err := json.Marshal(config.Config)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(b, etcd)
		if err != nil {
			return nil, err
		}
		storage, err := etcd.Open(logger)
		if err != nil {
			return nil, err
		}
		return storage, nil
	}
	return nil, ErrUnknownStorageAdapter
}
