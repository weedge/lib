module github.com/weedge/lib

go 1.15

require (
	github.com/Shopify/sarama v1.38.1
	github.com/gin-gonic/gin v1.9.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/google/uuid v1.3.0
	github.com/huandu/skiplist v1.2.0
	github.com/ii64/gouring v0.4.1
	github.com/jonboulle/clockwork v0.3.0 // indirect
	github.com/json-iterator/go v1.1.12
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
	github.com/lestrrat-go/strftime v1.0.6 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.14.0
	github.com/spf13/viper v1.15.0
	github.com/stretchr/testify v1.8.2
	github.com/xdg-go/scram v1.1.2
	go.etcd.io/etcd/api/v3 v3.5.7
	go.etcd.io/etcd/client/v3 v3.5.7
	go.mongodb.org/mongo-driver v1.11.3
	go.uber.org/atomic v1.10.0
	go.uber.org/zap v1.24.0
	golang.org/x/sys v0.6.0
	google.golang.org/grpc v1.54.0
	google.golang.org/protobuf v1.30.0
)

replace github.com/ii64/gouring => github.com/weedge/gouring v0.0.0-20230406152517-60a6d6e09b3a

//replace github.com/ii64/gouring => ../gouring
