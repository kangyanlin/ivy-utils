package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	clientv3 "github.com/coreos/etcd/clientv3"
	uuid "github.com/satori/go.uuid"
	core "github.com/universonic/ivy-utils/pkg/storage/core"
	zap "go.uber.org/zap"
)

const (
	hostPrefix = "host"

	// defaultStorageTimeout will be applied to all storage's operations.
	defaultStorageTimeout = 5 * time.Second
)

type conn struct {
	db     *clientv3.Client
	logger *zap.SugaredLogger
}

func (c *conn) Close() error {
	return c.db.Close()
}

func (c *conn) CreateHost(host core.Host) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultStorageTimeout)
	defer cancel()
	if _, err := uuid.FromString(host.GUID); err != nil {
		host.GUID = uuid.NewV4().String()
	}
	// NOTE: we are currently using hostname as host's primary unique identifier
	return c.txnCreate(ctx, canonicalID(hostPrefix, host.Hostname), host)
}

func (c *conn) GetHost(id string) (host core.Host, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultStorageTimeout)
	defer cancel()
	if err = c.getKey(ctx, canonicalID(hostPrefix, id), &host); err != nil {
		return
	}
	return host, nil
}

func (c *conn) UpdateHost(id string, updater func(host core.Host) (core.Host, error)) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultStorageTimeout)
	defer cancel()
	return c.txnUpdate(ctx, canonicalID(hostPrefix, id), func(currentValue []byte) ([]byte, error) {
		current := core.NewHost()
		if len(currentValue) > 0 {
			if err := json.Unmarshal(currentValue, current); err != nil {
				return nil, err
			}
		}
		updated, err := updater(*current)
		if err != nil {
			return nil, err
		}
		if _, err := uuid.FromString(updated.GUID); err != nil {
			updated.GUID = uuid.NewV4().String()
		}
		return json.Marshal(updated)
	})
}

func (c *conn) DeleteHost(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultStorageTimeout)
	defer cancel()
	return c.deleteKey(ctx, canonicalID(hostPrefix, id))
}

func (c *conn) ListHost() (hosts []core.Host, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultStorageTimeout)
	defer cancel()
	res, err := c.db.Get(ctx, hostPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	for _, v := range res.Kvs {
		var host core.Host
		if err = json.Unmarshal(v.Value, &host); err != nil {
			return nil, err
		}
		hosts = append(hosts, host)
	}
	return hosts, nil
}

func (c *conn) txnCreate(ctx context.Context, key string, value interface{}) (err error) {
	defer func() {
		defer c.logger.Sync()
		if err != nil {
			c.logger.Errorf("Error occurred during creating data entity '%s' due to: %v", key, err)
		} else {
			c.logger.Debugf("Created key '%s': %v", key, value)
		}
	}()
	var b []byte
	b, err = json.Marshal(value)
	if err != nil {
		return err
	}
	txn := c.db.Txn(ctx)
	var res *clientv3.TxnResponse
	res, err = txn.
		If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, string(b))).
		Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return core.ErrResourceAlreadyExists
	}
	return nil
}

func (c *conn) getKey(ctx context.Context, key string, value interface{}) (err error) {
	defer func() {
		defer c.logger.Sync()
		if err != nil {
			c.logger.Errorf("Error occurred during getting data entity '%s' due to: %v", key, err)
		} else {
			c.logger.Debugf("Retrieved key '%s': %v", key, value)
		}
	}()
	var r *clientv3.GetResponse
	r, err = c.db.Get(ctx, key)
	if err != nil {
		return err
	}
	if r.Count == 0 {
		return core.ErrResourceNotFound
	}
	return json.Unmarshal(r.Kvs[0].Value, value)
}

func (c *conn) txnUpdate(ctx context.Context, key string, update func(current []byte) ([]byte, error)) (err error) {
	var updatedValue []byte
	defer func() {
		defer c.logger.Sync()
		if err != nil {
			c.logger.Errorf("Error occurred during updating data entity '%s' due to: %v", key, err)
		} else {
			c.logger.Debugf("Updated key '%s': %s", key, updatedValue)
		}
	}()
	var getResp *clientv3.GetResponse
	getResp, err = c.db.Get(ctx, key)
	if err != nil {
		return err
	}
	var currentValue []byte
	var modRev int64
	if len(getResp.Kvs) > 0 {
		currentValue = getResp.Kvs[0].Value
		modRev = getResp.Kvs[0].ModRevision
	}

	updatedValue, err = update(currentValue)
	if err != nil {
		return err
	}

	txn := c.db.Txn(ctx)
	var updateResp *clientv3.TxnResponse
	updateResp, err = txn.
		If(clientv3.Compare(clientv3.ModRevision(key), "=", modRev)).
		Then(clientv3.OpPut(key, string(updatedValue))).
		Commit()
	if err != nil {
		return err
	}
	if !updateResp.Succeeded {
		return fmt.Errorf("Could not update key=%q due to: concurrent conflicting update happened", key)
	}
	return nil
}

func (c *conn) deleteKey(ctx context.Context, key string) (err error) {
	defer func() {
		defer c.logger.Sync()
		if err != nil {
			c.logger.Errorf("Error occurred during deleting data entity '%s' due to: %v", key, err)
		} else {
			c.logger.Debugf("Deleted key '%s'", key)
		}
	}()
	var res *clientv3.DeleteResponse
	res, err = c.db.Delete(ctx, key)
	if err != nil {
		return err
	}
	if res.Deleted == 0 {
		return core.ErrResourceNotFound
	}
	return nil
}

func canonicalID(prefix, id string) string { return filepath.Join(prefix, id) }
