package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var (
	logger       *zap.Logger
	labelNames   = []string{"cert_hostname", "alert_email"}
	notAfter     = NewGauge("cert_not_after", "cert not after X Unix Epoch seconds ", labelNames)
	hostQueueLen = NewGauge("host_queue_length", "how many hosts in queue waiting to be check", []string{"address"})
	config       Config
	hostQueue    = make(chan Host, 10)
)

func init() {
	logger, _ = zap.NewProduction()
}

var configFile = flag.String("config", "config.toml", "配置文件")

func NewGauge(name, help string, labels []string) *prometheus.GaugeVec {
	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "exporter",
		Name:      name,
		Help:      help,
	}, labels)
	prometheus.MustRegister(gauge)
	return gauge
}

func main() {
	flag.Parse()
	defer logger.Sync()

	config = parseConfig()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	http.Handle("/metrics", promhttp.Handler())
	httpServer := http.Server{
		Addr: config.ListenAddress,
	}

	// waiting to exit
	go func(ctx context.Context) {
		<-signalCh
		cancel()
		httpServer.Shutdown(ctx)
	}(ctx)

	go runCollectHosts(ctx, time.Duration(config.RefreshIntervalSecond)*time.Second)

	go processHosts(ctx)

	logger.Info("http server started", zap.String("address", config.ListenAddress),
		zap.Int("ssl check concurrency", config.Concurrency))
	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		logger.Fatal("HTTP server ListenAndServe Error: %v", zap.Error(err))
	}

	logger.Info("Bye")
}
