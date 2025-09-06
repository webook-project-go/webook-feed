package ioc

import (
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	resolver2 "google.golang.org/grpc/resolver"
)

func InitEtcd() *clientv3.Client {
	addrs := viper.GetStringSlice("etcd.addrs")
	client, err := clientv3.New(clientv3.Config{
		Endpoints: addrs,
	})
	if err != nil {
		panic(err)
	}
	return client
}
func InitResolver(client *clientv3.Client) resolver2.Builder {
	bd, err := resolver.NewBuilder(client)
	if err != nil {
		panic(err)
	}
	return bd
}
