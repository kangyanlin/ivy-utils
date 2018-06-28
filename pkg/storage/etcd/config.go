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

package etcd

import (
	"path/filepath"
	"time"

	clientv3 "github.com/coreos/etcd/clientv3"
	namespace "github.com/coreos/etcd/clientv3/namespace"
	transport "github.com/coreos/etcd/pkg/transport"
	core "github.com/universonic/ivy-utils/pkg/storage/core"
	zap "go.uber.org/zap"
)

const (
	defaultDialTimeout = 2 * time.Second
)

type EtcdSSLOptions struct {
	Enabled    bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	ServerName string `json:"server_name,omitempty" yaml:"server_name,omitempty"`
	CACert     string `json:"ca_cert,omitempty" yaml:"ca_cert,omitempty"`
	SSLKey     string `json:"key,omitempty" yaml:"key,omitempty"`
	SSLCert    string `json:"cert,omitempty" yaml:"cert,omitempty"`
}

type Etcd struct {
	Endpoints  []string        `json:"endpoints,omitempty" yaml:"endpoints,omitempty"`
	Namespace  []string        `json:"-" yaml:"-"`
	User       string          `json:"user,omitempty" yaml:"user,omitempty"`
	Password   string          `json:"password,omitempty" yaml:"password,omitempty"`
	SSLOptions *EtcdSSLOptions `json:"ssl,omitempty" yaml:"ssl,omitempty"`
}

func (in *Etcd) Open(logger *zap.SugaredLogger) (core.Storage, error) {
	cfg := clientv3.Config{
		Endpoints:   in.Endpoints,
		DialTimeout: defaultDialTimeout * time.Second,
		Username:    in.User,
		Password:    in.Password,
	}

	var cfgtls *transport.TLSInfo
	tlsinfo := transport.TLSInfo{}
	if in.SSLOptions.SSLCert != "" {
		tlsinfo.CertFile = in.SSLOptions.SSLCert
		cfgtls = &tlsinfo
	}

	if in.SSLOptions.SSLKey != "" {
		tlsinfo.KeyFile = in.SSLOptions.SSLKey
		cfgtls = &tlsinfo
	}

	if in.SSLOptions.CACert != "" {
		tlsinfo.CAFile = in.SSLOptions.CACert
		cfgtls = &tlsinfo
	}

	if in.SSLOptions.ServerName != "" {
		tlsinfo.ServerName = in.SSLOptions.ServerName
		cfgtls = &tlsinfo
	}

	if cfgtls != nil {
		clientTLS, err := cfgtls.ClientConfig()
		if err != nil {
			return nil, err
		}
		cfg.TLS = clientTLS
	}

	cfg.DialTimeout = 3 * time.Second

	db, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}
	in.Namespace = []string{"/cn.ivyent", "ivy-utils"}
	if len(in.Namespace) > 0 {
		db.KV = namespace.NewKV(db.KV, filepath.Join(in.Namespace...))
	}
	c := &conn{
		db:     db,
		logger: logger,
	}
	return c, nil
}

func New() *Etcd {
	return &Etcd{
		SSLOptions: new(EtcdSSLOptions),
	}
}
