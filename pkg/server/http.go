package server

import (
	"github.com/gin-gonic/gin"
	"github.com/openinsight-proj/elastic-alert/pkg/conf"
	"github.com/openinsight-proj/elastic-alert/pkg/server/controller"
	"github.com/openinsight-proj/elastic-alert/pkg/utils/logger"
)

func InitHttpServer(serverConfig *conf.AppConfig) {
	if serverConfig.Server.Enabled {
		router := NewRouter()
		err := router.Run(serverConfig.Server.ListenAddr)
		if err != nil {
			logger.Logger.Errorf("init http server failed: %s", err.Error())
		}
	}
}

func NewRouter() *gin.Engine {
	r := gin.Default()

	v1Route := r.Group("/v1")

	{
		//TODO:
		v1Route.GET("/rules", controller.FindAllRules)
	}

	return r
}
