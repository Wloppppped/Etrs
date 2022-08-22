package internal

import (
	"Etrs/common"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

const (
	TTL_KEY     = "ETRS_ENV_TTL_KEY"
	DEFAULT_TTL = 60
)

type EtrsRegistry struct {
	etcdClient *clientv3.Client
	leaseTTL   int64
	meta       *registerMeta
}

type registerMeta struct {
	lease         clientv3.Lease
	leaseID       clientv3.LeaseID
	ctx           context.Context
	cancel        context.CancelFunc
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
}

// EtrsRegistry creates a etcd based registry.
func EtrsRegistryInit(endpoints []string, opts ...Option) (*EtrsRegistry, error) {
	cfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	etcdClient, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}
	er := &EtrsRegistry{
		etcdClient: etcdClient,
		leaseTTL:   getTTL(),
	}

	// check
	if er.etcdClient == nil {
		return nil, errors.New("wrong Etcd client")
	}

	er.setTTL()

	go er.watchLease()
	return er, nil
}

func (er *EtrsRegistry) setTTL() error {
	client := er.etcdClient
	lease := clientv3.NewLease(client)

	//设置租约时间
	leaseResp, err := lease.Grant(context.TODO(), er.leaseTTL)
	if err != nil {
		return err
	}

	ctx, cancelFunc := context.WithCancel(context.TODO())
	leaseRespChan, err := lease.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		cancelFunc()
		return err
	}

	er.meta = &registerMeta{}
	er.meta.cancel = cancelFunc
	er.meta.ctx = ctx
	er.meta.lease = lease
	er.meta.leaseID = leaseResp.ID
	er.meta.keepAliveChan = leaseRespChan
	zap.L().Debug("set lease ttl success!")
	return nil
}

func (er *EtrsRegistry) watchLease() {
	watch := true
	for watch {
		resp := <-er.meta.keepAliveChan
		if resp == nil {
			zap.L().Info("keepAlive lease stop!")
			watch = false
		} else {
			zap.L().Info("keepAlive lease!")
		}
	}
}

func (er *EtrsRegistry) ServerRegistry(key, val string) error {
	kv := clientv3.NewKV(er.etcdClient)
	v, err := kv.Put(context.TODO(), fmt.Sprintf(common.ETCD_PREFIX_FMT, key), val, clientv3.WithLease(er.meta.leaseID))
	zap.L().Debug(fmt.Sprintf("server registry key: %s, value: %s, version: %v", key, val, v))
	return err
}

func (er *EtrsRegistry) ServerCancle() error {
	er.meta.cancel()
	// 等待一下
	time.Sleep(2 * time.Second)
	_, err := er.meta.lease.Revoke(context.TODO(), er.meta.leaseID)
	return err
}

func getTTL() int64 {
	var ttl int64 = DEFAULT_TTL
	if str, ok := os.LookupEnv(TTL_KEY); ok {
		if t, err := strconv.Atoi(str); err == nil {
			ttl = int64(t)
		}
	}
	return ttl
}
