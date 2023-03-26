package discovery

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/weedge/lib/runtimer"

	"github.com/weedge/lib/log"

	"github.com/pkg/errors"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

type GrpcResolver struct {
	*EtcdClient
	store map[string]map[string]struct{}
}

func NewGrpcResolver(endpoints []string, dialTimeout time.Duration) (m *GrpcResolver) {
	if len(endpoints) == 0 {
		endpoints = []string{"0.0.0.0:2379"}
	}
	if dialTimeout <= 0 {
		dialTimeout = 3 * time.Second
	}
	m = &GrpcResolver{
		EtcdClient: &EtcdClient{
			endpoints:   endpoints,
			dialTimeout: dialTimeout,
		},
		store: make(map[string]map[string]struct{}),
	}

	return
}

func (m *GrpcResolver) InitGrpcResolver() (err error) {
	m.client, err = clientv3.New(clientv3.Config{Endpoints: m.endpoints, DialTimeout: m.dialTimeout})
	if err != nil {
		log.Error(err)
		return
	}

	resolver.Register(m)

	return
}

func (b *GrpcResolver) DebugStore() {
	fmt.Printf("store %+v\n", b.store)
}

// Build creates a new resolver for the given target.
//
// gRPC dial calls Build synchronously, and fails if the returned error is
// not nil.
func (b *GrpcResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	fmt.Printf("call builder %s\n", target.Endpoint)
	b.store[target.Endpoint()] = make(map[string]struct{})

	r := &etcdResolver{
		client: b.client,
		target: target,
		cc:     cc,
		store:  b.store[target.Endpoint()],
		stopCh: make(chan struct{}, 1),
		rn:     make(chan struct{}, 1),
		t:      time.NewTicker(defaultFreq),
	}

	runtimer.GoSafely(nil, false, func() {
		r.start(context.Background())
	}, nil, nil)
	r.ResolveNow(resolver.ResolveNowOptions{})

	return r, nil
}

// Scheme returns the scheme supported by this resolver.
// Scheme is defined at https://github.com/grpc/grpc/blob/master/doc/naming.md.
func (b *GrpcResolver) Scheme() string {
	return Scheme
}

type etcdResolver struct {
	client *clientv3.Client
	target resolver.Target
	cc     resolver.ClientConn
	store  map[string]struct{}
	stopCh chan struct{}
	rn     chan struct{} // rn channel is used by ResolveNow() to force an immediate resolution of the target.
	t      *time.Ticker
}

func (r *etcdResolver) start(ctx context.Context) {
	target := r.target.Endpoint()

	w := clientv3.NewWatcher(r.client)
	rch := w.Watch(ctx, target+"/", clientv3.WithPrefix())
	for {
		select {
		case <-r.rn:
			r.resolveNow()
		case <-r.t.C:
			r.ResolveNow(resolver.ResolveNowOptions{})
		case <-r.stopCh:
			err := w.Close()
			if err != nil {
				log.Errorf("etcd watcher close err", err.Error())
			}
			return
		case wresp := <-rch:
			for _, ev := range wresp.Events {
				switch ev.Type {
				case mvccpb.PUT:
					r.store[string(ev.Kv.Value)] = struct{}{}
					log.Info("etcd watcher put kv", string(ev.Kv.Key), string(ev.Kv.Value))
				case mvccpb.DELETE:
					delete(r.store, strings.Replace(string(ev.Kv.Key), target+"/", "", 1))
					log.Info("etcd watcher delete kv", string(ev.Kv.Key), string(ev.Kv.Value))
				}
			}
			r.updateTargetState()
		}
	}
}

func (r *etcdResolver) resolveNow() {
	target := r.target.Endpoint()
	resp, err := r.client.Get(context.Background(), target+"/", clientv3.WithPrefix())
	if err != nil {
		r.cc.ReportError(errors.Wrap(err, "get init endpoints"))
		return
	}

	for _, kv := range resp.Kvs {
		r.store[string(kv.Value)] = struct{}{}
	}

	r.updateTargetState()
}

func (r *etcdResolver) updateTargetState() {
	addrs := make([]resolver.Address, len(r.store))
	i := 0
	for k := range r.store {
		addrs[i] = resolver.Address{Addr: k}
		i++
	}
	r.cc.UpdateState(resolver.State{Addresses: addrs})
}

// ResolveNow will be called by gRPC to try to resolve the target name
// again. It's just a hint, resolver can ignore this if it's not necessary.
//
// It could be called multiple times concurrently.
func (r *etcdResolver) ResolveNow(o resolver.ResolveNowOptions) {
	select {
	case r.rn <- struct{}{}:
	default:

	}
}

// Close closes the resolver.
func (r *etcdResolver) Close() {
	r.t.Stop()
	close(r.stopCh)
}
