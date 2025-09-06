//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/webook-project-go/webook-feed/grpc"
	"github.com/webook-project-go/webook-feed/ioc"
	"github.com/webook-project-go/webook-feed/repository"
	"github.com/webook-project-go/webook-feed/repository/cache"
	"github.com/webook-project-go/webook-feed/repository/dao"
	"github.com/webook-project-go/webook-feed/service"
)

var thirdPartyProvider = wire.NewSet(
	ioc.InitDatabase,
	ioc.InitRedis,
	ioc.InitKafka,
	ioc.InitLogger,
	ioc.InitConsumer,
	ioc.InitEtcd,
	ioc.InitActive,
	ioc.InitRelation,
	ioc.InitGrpcServer,
	ioc.InitResolver,
)
var feedServiceSet = wire.NewSet(
	service.NewService,
	repository.NewRepository,
	cache.NewCache,
	dao.NewDao,
)

func InitApp() *App {
	wire.Build(
		wire.Struct(new(App), "*"),
		thirdPartyProvider,
		feedServiceSet,
		grpc.NewService,
	)
	return new(App)
}
