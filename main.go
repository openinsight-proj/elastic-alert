package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/openinsight-proj/elastic-alert/pkg/client"
	"github.com/openinsight-proj/elastic-alert/pkg/server"

	"github.com/jessevdk/go-flags"
	"github.com/openinsight-proj/elastic-alert/pkg/boot"
	"github.com/openinsight-proj/elastic-alert/pkg/conf"
	"github.com/openinsight-proj/elastic-alert/pkg/utils/logger"
	"github.com/openinsight-proj/elastic-alert/pkg/utils/redis"
	"github.com/openinsight-proj/elastic-alert/pkg/utils/xtime"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	var opts conf.FlagOption

	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("%s\n", e)
			if opts.Debug {
				debug.PrintStack()
			}
		}
	}()

	p := flags.NewParser(&opts, flags.HelpFlag)
	_, err := p.ParseArgs(os.Args)
	if err != nil {
		panic(err)
	}

	logger.SetLogLevel(opts.GetLogLevel())
	xtime.FixedZone(opts.Zone)

	c := conf.GetAppConfig(opts.ConfigPath)

	var ea *boot.ElasticAlert

	run := func(ctx context.Context, kubeClient *client.KubeClient, c *conf.AppConfig) {
		// complete your controller loop here
		klog.Info("Controller loop...")

		// only set up redis when alertmanager enabled.
		if c.Alert.Alertmanager.Enabled {
			redis.Setup()
		}

		ea = boot.NewElasticAlert(c, &opts)
		ea.Start()
		startHttpServer(c, ea, kubeClient)

		if c.Exporter.Enabled {
			metrics := boot.NewRuleStatusCollector(ea)
			reg := prometheus.NewPedanticRegistry()
			err := reg.Register(metrics)
			if err != nil {
				t := fmt.Sprintf("Register prometheus collector error: %s", err.Error())
				panic(t)
			}
			gatherers := prometheus.Gatherers{
				prometheus.DefaultGatherer,
				reg,
			}
			h := promhttp.HandlerFor(gatherers,
				promhttp.HandlerOpts{
					ErrorHandling: promhttp.ContinueOnError,
				})
			http.Handle("/metrics", h)
			http.HandleFunc("/alert/message", boot.RenderAlertMessage)
			e := http.ListenAndServe(c.Exporter.ListenAddr, nil)

			if e != nil {
				t := fmt.Sprintf("Prometheus exporter start error: %s", e.Error())
				panic(t)
			}
		}
	}

	var kube_client *client.KubeClient
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if c.Server.Enabled {
		kube_client, err = client.NewKubeClient(c.Server.DB.Config)
		if err != nil {
			logger.Logger.Errorf("could not get kube clientset: %v.", err)
		}

		if c.Server.LeaderElection.Enabled {
			id, err := os.Hostname()
			if err != nil || id == "" {
				id = fmt.Sprintf("elastic-alert-%s", uuid.New().String())
			}

			// we use the Lease lock type since edits to Leases are less common
			// and fewer objects in the cluster watch "all Leases".
			lock := &resourcelock.LeaseLock{
				LeaseMeta: metav1.ObjectMeta{
					Name:      "elastic-alert",
					Namespace: c.Server.LeaderElection.Namespace,
				},
				Client: kube_client.Client.CoordinationV1(),
				LockConfig: resourcelock.ResourceLockConfig{
					Identity: id,
				},
			}

			// start the leader election code loop
			leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
				Lock:            lock,
				ReleaseOnCancel: true,
				LeaseDuration:   60 * time.Second,
				RenewDeadline:   15 * time.Second,
				RetryPeriod:     5 * time.Second,
				Callbacks: leaderelection.LeaderCallbacks{
					OnStartedLeading: func(ctx context.Context) {
						run(ctx, kube_client, c)
					},
					OnStoppedLeading: func() {
						// if lose leader, sleep 3 minutes and try to became leader again
						klog.Info("lose lost: %s, will sleep three minutes before retry", id)
						time.Sleep(3 * time.Minute)
					},
					OnNewLeader: func(identity string) {
						if identity == id {
							return
						}
						klog.Infof("new leader elected: %s", identity)
					},
				},
			})
		} else {
			run(ctx, kube_client, c)
		}
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	for {
		s := <-quit
		switch s {
		case syscall.SIGHUP:
			c := conf.GetAppConfig(opts.ConfigPath)
			ea.SetAppConf(c)
			logger.Logger.Infoln("Reload application config success!")
		case syscall.SIGINT:
			fallthrough
		case syscall.SIGTERM:
			ea.Stop()
			logger.Logger.Infoln("exiting...")
			return
		}
	}
}

func startHttpServer(conf *conf.AppConfig, ea *boot.ElasticAlert, kube_client *client.KubeClient) {
	//init http server
	s := server.HttpServer{
		ServerConfig: conf,
		Ea:           ea,
		KubeClient:   kube_client,
	}
	go s.InitHttpServer()
}
