module github.com/opencord/voltctl

go 1.16

replace (
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.25.1
)

require (
	github.com/Shopify/sarama v1.29.1
	github.com/golang/protobuf v1.5.3
	github.com/googleapis/gnostic v0.3.0 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/jessevdk/go-flags v1.4.0
	github.com/jhump/protoreflect v1.10.2
	github.com/opencord/voltha-lib-go/v7 v7.6.3
	github.com/opencord/voltha-protos/v5 v5.6.5
	github.com/stretchr/testify v1.8.1
	google.golang.org/grpc v1.56.2
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.0.0-20190819141258-3544db3b9e44
	k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/client-go v0.0.0-20190819141724-e14f31a72a77 // go get k8s.io/client-go@kubernetes-1.15.3
)
