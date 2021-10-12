package discovery

import (
	"time"

	"go.etcd.io/etcd/client/v3"
)

const (
	DEBUG       = false
	Scheme      = "etcdv3"
	defaultFreq = time.Minute * 30
)

type EtcdClient struct {
	endpoints   []string
	dialTimeout time.Duration
	client      *clientv3.Client
}


