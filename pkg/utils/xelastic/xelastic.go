package xelastic

import (
	"crypto/tls"
	"net/http"

	"github.com/openinsight-proj/elastic-alert/pkg/model"

	elasticsearch8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/openinsight-proj/elastic-alert/pkg/utils/logger"
)

type ElasticClient interface {
	FindByDSL(index string, dsl string, source []string) ([]any, int, int)
	CountByDSL(index string, dsl string) (int, int)
}

func NewElasticClient(esConfig model.EsConfig, version string) ElasticClient {
	client, err := elasticsearch8.NewClient(elasticsearch8.Config{
		Addresses: esConfig.Addresses,
		Username:  esConfig.Username,
		Password:  esConfig.Password,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	})
	if err != nil {
		logger.Logger.Errorln(err)
		return nil
	}
	c := &ElasticClientV8{
		client: client,
	}
	return c
}
