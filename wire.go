//go:build wireinject

package main

import (
	"github.com/dadaxiaoxiao/go-pkg/customserver"
	"github.com/dadaxiaoxiao/tag/internal/grpc"
	"github.com/dadaxiaoxiao/tag/internal/repository/cache"
	"github.com/dadaxiaoxiao/tag/internal/repository/dao"
	"github.com/dadaxiaoxiao/tag/internal/service"
	"github.com/dadaxiaoxiao/tag/ioc"
	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(
	ioc.InitEtcdClient,
	ioc.InitLogger,
	ioc.InitRedis,
	ioc.InitDB,
	ioc.InitKafka,
	ioc.InitProducer,
)

func InitApp() *customserver.App {
	wire.Build(thirdProvider,
		dao.NewGORMTagDAO,
		cache.NewRedisTagCache,
		ioc.InitRepository,
		service.NewTagService,
		grpc.NewTagServiceServer,
		ioc.InitGRPCServer,
		wire.Struct(new(customserver.App), "GRPCServer"),
	)
	return new(customserver.App)
}
