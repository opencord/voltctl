module github.com/opencord/voltctl

go 1.16

replace (
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.25.1
)

require (
	github.com/Shopify/sarama v1.29.1
	github.com/golang/protobuf v1.5.2
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/jessevdk/go-flags v1.4.0
	github.com/jhump/protoreflect v1.5.0
	github.com/opencord/voltha-lib-go/v7 v7.1.0
	github.com/opencord/voltha-protos/v5 v5.4.10
	github.com/stretchr/testify v1.7.0
	google.golang.org/grpc v1.44.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.20.0-alpha.2
	k8s.io/apimachinery v0.20.0-alpha.2
	k8s.io/client-go v0.20.0-alpha.2 // go get k8s.io/client-go@kubernetes-1.15.3
)
