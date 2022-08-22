package internal

import (
	"Etrs/common"
	"context"
	"fmt"
	"sync"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type EtrsResolver struct {
	etcdClient       *clientv3.Client
	specialAttention map[string]*valueEntry
}

type valueEntry struct {
	value   string
	version int64
}

var localCacheLock sync.RWMutex

// EtrsResolverInit creates a etcd based resolver.
func EtrsResolverInit(endpoints []string, opts ...Option) (*EtrsResolver, error) {
	cfg := clientv3.Config{
		Endpoints: endpoints,
	}
	// opts
	for _, opt := range opts {
		opt(&cfg)
	}
	etcdClient, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}
	return &EtrsResolver{
		etcdClient:       etcdClient,
		specialAttention: map[string]*valueEntry{},
	}, nil
}

func (er *EtrsResolver) GetServicePrefix(prefix string) ([]string, error) {
	signal := make(chan struct{})
	go er.watcher(prefix, signal)
	// Ensure successful watch
	<-signal

	resp, err := er.etcdClient.Get(context.Background(), fmt.Sprintf(common.ETCD_PREFIX_FMT, prefix), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	addrs := er.extractAddrs(resp)

	return addrs, nil
}

func (er *EtrsResolver) watcher(prefix string, signal chan struct{}) {
	rch := er.etcdClient.Watch(context.Background(), fmt.Sprintf(common.ETCD_PREFIX_FMT, prefix), clientv3.WithPrefix())
	signal <- struct{}{}
	// range watch
	for wresp := range rch {
		// each channel may have multiple events
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT:
				er.setServiceList(ev.Kv)
			case mvccpb.DELETE:
				er.delServiceList(ev.Kv)
			}
		}
	}
}

func (er *EtrsResolver) extractAddrs(resp *clientv3.GetResponse) []string {
	addrs := make([]string, 0)
	if resp == nil || resp.Kvs == nil {
		return addrs
	}
	for i := range resp.Kvs {
		if v := resp.Kvs[i].Value; v != nil {
			er.setServiceList(resp.Kvs[i])
			addrs = append(addrs, string(v))
		}
	}
	return addrs
}

func (er *EtrsResolver) setServiceList(kv *mvccpb.KeyValue) {
	localCacheLock.Lock()
	defer localCacheLock.Unlock()

	k := kv.Key
	v := kv.Value
	version := kv.Version

	old, exixt := er.specialAttention[string(k)]
	// 如果不存在 || kvVersion 新于 cacheVersion
	if !exixt || old.version < version {
		er.specialAttention[string(k)] = &valueEntry{
			value:   string(v),
			version: version,
		}
	}
	if exixt {
		zap.L().Debug(fmt.Sprintf("local cache store k %s, v %s, i %v, o %v", k, v, version, old.version))
	} else {
		zap.L().Debug(fmt.Sprintf("local cache store k %s, v %s, i %v", k, v, version))
	}

}

func (er *EtrsResolver) delServiceList(kv *mvccpb.KeyValue) {
	localCacheLock.Lock()
	defer localCacheLock.Unlock()

	k := kv.Key
	delete(er.specialAttention, string(k))
	zap.L().Debug(fmt.Sprintf("local cache delete k %s", k))
}

func (er *EtrsResolver) Resolver(k string) (string, bool) {
	localCacheLock.RLock()
	defer localCacheLock.RUnlock()
	v, ok := er.specialAttention[fmt.Sprintf(common.ETCD_PREFIX_FMT, k)]
	if !ok {
		return "", ok
	}
	return v.value, true
}
