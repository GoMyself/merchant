module merchant

go 1.18

replace github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

require (
	github.com/apache/rocketmq-client-go/v2 v2.1.0
	github.com/beanstalkd/go-beanstalk v0.1.0
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869
	github.com/coreos/etcd v3.3.27+incompatible
	github.com/doug-martin/goqu/v9 v9.18.0
	github.com/fasthttp/router v1.4.9
	github.com/fluent/fluent-logger-golang v1.9.0
	github.com/go-redis/redis/v8 v8.11.4
	github.com/go-sql-driver/mysql v1.6.0
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	github.com/goccy/go-json v0.9.4
	github.com/hprose/hprose-golang/v3 v3.0.8
	github.com/ip2location/ip2location-go/v9 v9.4.0
	github.com/ipipdotnet/ipdb-go v1.3.1
	github.com/jmoiron/sqlx v1.3.4
	github.com/json-iterator/go v1.1.12
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/meilisearch/meilisearch-go v0.19.1
	github.com/minio/md5-simd v1.1.2
	github.com/modern-go/reflect2 v1.0.2
	github.com/nats-io/nats.go v1.13.1-0.20211122170419-d7c1d78a50fc
	github.com/olivere/elastic/v7 v7.0.31
	github.com/panjf2000/ants/v2 v2.4.7
	github.com/pelletier/go-toml v1.9.4
	github.com/shopspring/decimal v1.3.1
	github.com/silenceper/pool v1.0.0
	github.com/spaolacci/murmur3 v1.1.0
	github.com/taosdata/driver-go/v2 v2.0.1-0.20220512023129-15f5b9c4b11c
	github.com/tinylib/msgp v1.1.6
	github.com/valyala/fasthttp v1.36.0
	github.com/valyala/fastjson v1.6.3
	github.com/wI2L/jettison v0.7.3
	github.com/xxtea/xxtea-go v0.0.0-20170828040851-35c4b17eecf6
	go.uber.org/automaxprocs v1.4.0
	lukechampine.com/frand v1.4.2
)

require (
	github.com/aead/chacha20 v0.0.0-20180709150244-8b13a72661da // indirect
	github.com/andot/complexconv v1.0.0 // indirect
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/coreos/bbolt v0.0.0-00010101000000-000000000000 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/mock v1.4.4 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.15.0 // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/kr/text v0.1.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/nats-io/nats-server/v2 v2.6.6 // indirect
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/philhofer/fwd v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.12.0 // indirect
	github.com/savsgio/gotils v0.0.0-20220401102855-e56b59f40436 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/soheilhy/cmux v0.1.5 // indirect
	github.com/tidwall/gjson v1.2.1 // indirect
	github.com/tidwall/match v1.0.1 // indirect
	github.com/tidwall/pretty v0.0.0-20190325153808-1166b9ac2b65 // indirect
	github.com/tmc/grpc-websocket-proxy v0.0.0-20220101234140-673ab2c3ae75 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.20.0 // indirect
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sys v0.0.0-20220227234510-4e6760a101f9 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20211116232009-f0f3c7e86c11 // indirect
	google.golang.org/genproto v0.0.0-20200825200019-8632dd797987 // indirect
	google.golang.org/grpc v1.33.2 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
	stathat.com/c/consistent v1.0.0 // indirect
)
