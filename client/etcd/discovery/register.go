package discovery

import (
	"context"
	"strings"
	"time"

	"github.com/weedge/lib/log"
	"github.com/weedge/lib/runtimer"

	"github.com/pkg/errors"
	"go.etcd.io/etcd/client/v3"
)

type Register struct {
	*EtcdClient
	serviceName string
	selfIpPort  string
	cancel      context.CancelFunc
}

func NewRegister(endpoints []string, dialTimeout time.Duration, serviceName, selfIpPort string) (m *Register) {
	if len(endpoints) == 0 {
		endpoints = []string{"0.0.0.0:2379"}
	}
	if dialTimeout <= 0 {
		dialTimeout = 3 * time.Second
	}

	m = &Register{
		serviceName: serviceName,
		selfIpPort:  selfIpPort,
		EtcdClient: &EtcdClient{
			endpoints:   endpoints,
			dialTimeout: dialTimeout,
		},
	}

	return
}

func (m *Register) Do() (err error) {
	m.client, err = clientv3.New(clientv3.Config{Endpoints: m.endpoints, DialTimeout: m.dialTimeout})
	if err != nil {
		log.Error(err)
		return
	}

	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	runtimer.GoSafely(nil, false, func() {
		err = m.register(ctx)
		if err != nil {
			log.Error(err)
			return
		}
	}, nil, nil)

	return
}

func (m *Register) register(ctx context.Context) error {
	resp, err := m.client.Grant(ctx, 5)
	if err != nil {
		return errors.Wrap(err, "etcd grant")
	}
	_, err = m.client.Put(ctx, strings.Join([]string{m.serviceName, m.selfIpPort}, "/"), m.selfIpPort, clientv3.WithLease(resp.ID))
	if err != nil {
		return errors.Wrap(err, "etcd put")
	}
	log.Infof("etcd put ok; path:%s selfIpPort:%s leaseId:%d",
		strings.Join([]string{m.serviceName, m.selfIpPort}, "/"), m.selfIpPort, resp.ID)
	respCh, err := m.client.KeepAlive(ctx, resp.ID)
	if err != nil {
		return errors.Wrap(err, "etcd keep alive")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case res, ok := <-respCh:
			if ok {
				if DEBUG {
					log.Infof("etcd keepalive resp:%+v", res)
				}
			}
		}
	}
}

func (m *Register) Close() {

	m.cancel()
}
