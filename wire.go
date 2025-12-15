//go:build wireinject

package main

import (
	"connectify/internal/repository"
	"connectify/internal/repository/cache"
	"connectify/internal/repository/dao"
	"connectify/internal/service"
	"connectify/internal/web"
	"connectify/ioc"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// Third-party dependencies
		ioc.InitRedis, ioc.InitDB,
		// DAO part
		dao.NewUserDao,

		// cache part
		cache.NewCodeCache, cache.NewUserCache,

		// repository part
		repository.NewUserRepository,
		repository.NewCodeRepository,

		// Service part
		ioc.InitSmsService,
		service.NewUserService,
		service.NewCodeService,

		// handler part
		web.NewUserHandler,

		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
