package server

import (
	"github.com/gin-gonic/gin"
	"github.com/openinsight-proj/elastic-alert/pkg/boot"
	"github.com/openinsight-proj/elastic-alert/pkg/conf"
	"github.com/openinsight-proj/elastic-alert/pkg/server/controller"
	"github.com/openinsight-proj/elastic-alert/pkg/utils/logger"
)

type HttpServer struct {
	ServerConfig *conf.AppConfig
	Ea           *boot.ElasticAlert
}

func (s *HttpServer) InitHttpServer() {
	if s.ServerConfig.Server.Enabled {
		router := s.newRouter()
		err := router.Run(s.ServerConfig.Server.ListenAddr)
		if err != nil {
			logger.Logger.Errorf("init http server failed: %s", err.Error())
		}
	}
}

func (s *HttpServer) newRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/reload", func(c *gin.Context) {
		s.Ea.Loader.ReloadSchedulerJob(s.Ea)
	})

	v1Route := r.Group("/v1")

	{
		//TODO:
		v1Route.GET("/rules", controller.FindAllRules)
	}

	return r
}