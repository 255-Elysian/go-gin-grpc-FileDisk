package discovery

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
	"time"
)

const (
	schema = "etcd"
)

type Resolver struct {
	schema      string
	EtcdAddrs   []string
	DialTimeout int

	closeCh      chan struct{}
	watchCh      clientv3.WatchChan
	cil          *clientv3.Client
	keyPrefix    string
	srvAddrsList []resolver.Address

	cc     resolver.ClientConn
	logger *zap.Logger
}

func NewResolver(etcdAddrs []string, logger *zap.Logger) *Resolver {
	return &Resolver{
		schema:      schema,
		EtcdAddrs:   etcdAddrs,
		DialTimeout: 3,
		logger:      logger,
	}
}

func (r *Resolver) Scheme() string {
	return r.schema
}

func (r *Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r.cc = cc

	r.keyPrefix = BuildPreFix(Server{Name: target.Endpoint(), Version: target.URL.Host})
	if _, err := r.start(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Resolver) ResolveNow(o resolver.ResolveNowOptions) {}

func (r *Resolver) Close() {
	r.closeCh <- struct{}{}
}

func (r *Resolver) start() (chan<- struct{}, error) {
	var err error
	r.cil, err = clientv3.New(clientv3.Config{
		Endpoints:   r.EtcdAddrs,
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	})
	if err != nil {
		return nil, err
	}
	resolver.Register(r)

	r.closeCh = make(chan struct{})

	if err = r.sync(); err != nil {
		return nil, err
	}

	go r.watch()

	return r.closeCh, nil
}

func (r *Resolver) watch() {
	ticker := time.NewTicker(time.Minute)
	r.watchCh = r.cil.Watch(context.Background(), r.keyPrefix, clientv3.WithPrefix())

	for {
		select {
		case <-r.closeCh:
			return
		case res, ok := <-r.watchCh:
			if ok {
				r.update(res.Events)
			}
		case <-ticker.C:
			if err := r.sync(); err != nil {
				r.logger.Error("sync failed, err:", zap.Error(err))
			}
		}
	}
}

func (r *Resolver) update(events []*clientv3.Event) {
	for _, event := range events {
		var err error
		var info Server

		switch event.Type {
		case clientv3.EventTypePut:
			info, err = ParseValue(event.Kv.Value)
			if err != nil {
				continue
			}
			addr := resolver.Address{Addr: info.Addr, Metadata: info.Weight}
			if !Exist(r.srvAddrsList, addr) {
				r.srvAddrsList = append(r.srvAddrsList, addr)
				err := r.cc.UpdateState(resolver.State{Addresses: r.srvAddrsList})
				if err != nil {
					return
				}
			}
		case clientv3.EventTypeDelete:
			info, err = ParseValue(event.Kv.Value)
			if err != nil {
				continue
			}
			addr := resolver.Address{Addr: info.Addr}
			if s, ok := Remove(r.srvAddrsList, addr); ok {
				r.srvAddrsList = s
				err := r.cc.UpdateState(resolver.State{Addresses: r.srvAddrsList})
				if err != nil {
					return
				}
			}
		}
	}
}

// sync 同步获取所有地址信息
func (r *Resolver) sync() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	resp, err := r.cil.Get(ctx, r.keyPrefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	r.srvAddrsList = []resolver.Address{}

	for _, v := range resp.Kvs {
		info, err := ParseValue(v.Value)
		if err != nil {
			continue
		}
		addr := resolver.Address{Addr: info.Addr, Metadata: info.Weight}
		r.srvAddrsList = append(r.srvAddrsList, addr)
	}
	err = r.cc.UpdateState(resolver.State{Addresses: r.srvAddrsList})
	if err != nil {
		return err
	}
	return nil
}
