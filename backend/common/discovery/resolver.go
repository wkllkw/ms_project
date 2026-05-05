package discovery

import (
	"context"
	"strings"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

const (
	schema = "etcd"
)

// Resolver for grpc client
type Resolver struct {
	schema      string
	EtcdAddrs   []string
	DialTimeout int

	closeCh      chan struct{}
	watchCh      clientv3.WatchChan
	cli          *clientv3.Client
	keyPrifix    string
	srvAddrsList []resolver.Address

	cc     resolver.ClientConn
	logger *zap.Logger
}

// NewResolver create a new resolver.Builder base on etcd
func NewResolver(etcdAddrs []string, logger *zap.Logger) *Resolver {
	return &Resolver{
		schema:      schema,
		EtcdAddrs:   etcdAddrs,
		DialTimeout: 3,
		logger:      logger,
	}
}

// Scheme returns the scheme supported by this resolver.
func (r *Resolver) Scheme() string {
	return r.schema
}

// Build creates a new resolver.Resolver for the given target
func (r *Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r.cc = cc

	// 优先使用 target.Endpoint（兼容 gRPC v1.51+ 的 deprecated 字段）
	// 如果为空则从 URL.Path 提取服务名
	endpoint := target.Endpoint
	if endpoint == "" {
		endpoint = strings.TrimPrefix(target.URL.Path, "/")
	}
	r.keyPrifix = BuildPrefix(Server{Name: endpoint, Version: target.Authority})
	if _, err := r.start(); err != nil {
		return nil, err
	}
	return r, nil
}

// ResolveNow resolver.Resolver interface
// 当 gRPC 连接遇到错误或需要重新解析时被调用，触发一次同步刷新
func (r *Resolver) ResolveNow(o resolver.ResolveNowOptions) {
	if r.cli == nil {
		return
	}
	if err := r.sync(); err != nil {
		r.logger.Error("ResolveNow sync failed", zap.Error(err))
	}
}

// Close resolver.Resolver interface
func (r *Resolver) Close() {
	r.closeCh <- struct{}{}
}

// start
func (r *Resolver) start() (chan<- struct{}, error) {
	var err error
	r.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.EtcdAddrs,
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	})
	if err != nil {
		return nil, err
	}
	resolver.Register(r)

	r.closeCh = make(chan struct{})

	// 首次同步：即使没有找到服务地址也不报错，等待后续 watch 或 ResolveNow 刷新
	// 解决服务启动顺序导致的 "produced zero addresses" 问题
	if err = r.sync(); err != nil {
		r.logger.Warn("initial sync found no addresses, will retry via watch/ResolveNow",
			zap.String("prefix", r.keyPrifix), zap.Error(err))
		// 不返回错误，让 watch 协程继续运行以等待服务注册
	}

	go r.watch()

	return r.closeCh, nil
}

// watch update events
func (r *Resolver) watch() {
	ticker := time.NewTicker(time.Minute)
	r.watchCh = r.cli.Watch(context.Background(), r.keyPrifix, clientv3.WithPrefix())

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
				r.logger.Error("sync failed", zap.Error(err))
			}
		}
	}
}

// update
func (r *Resolver) update(events []*clientv3.Event) {
	for _, ev := range events {
		var info Server
		var err error

		switch ev.Type {
		case mvccpb.PUT:
			info, err = ParseValue(ev.Kv.Value)
			if err != nil {
				continue
			}
			addr := resolver.Address{Addr: info.Addr, Metadata: info.Weight}
			if !Exist(r.srvAddrsList, addr) {
				r.srvAddrsList = append(r.srvAddrsList, addr)
				r.cc.UpdateState(resolver.State{Addresses: r.srvAddrsList})
			}
		case mvccpb.DELETE:
			info, err = SplitPath(string(ev.Kv.Key))
			if err != nil {
				continue
			}
			addr := resolver.Address{Addr: info.Addr}
			if s, ok := Remove(r.srvAddrsList, addr); ok {
				r.srvAddrsList = s
				r.cc.UpdateState(resolver.State{Addresses: r.srvAddrsList})
			}
		}
	}
}

// sync 同步获取所有地址信息
func (r *Resolver) sync() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	res, err := r.cli.Get(ctx, r.keyPrifix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	var addrs []resolver.Address
	for _, v := range res.Kvs {
		info, err := ParseValue(v.Value)
		if err != nil {
			continue
		}
		addr := resolver.Address{Addr: info.Addr, Metadata: info.Weight}
		addrs = append(addrs, addr)
	}

	// 仅在地址列表有变化时更新状态
	// 避免在没有服务实例时向 gRPC 报告空地址列表导致 "produced zero addresses" 错误
	if len(addrs) > 0 {
		r.srvAddrsList = addrs
		r.cc.UpdateState(resolver.State{Addresses: r.srvAddrsList})
	} else if len(r.srvAddrsList) > 0 {
		// 之前有地址，现在没有了（服务下线），需要更新
		r.srvAddrsList = addrs
		r.cc.UpdateState(resolver.State{Addresses: r.srvAddrsList})
	}
	// 首次同步且无地址：不更新状态，等待后续 watch 事件触发
	return nil
}
