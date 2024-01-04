package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var (
	logger     *zap.Logger
	labelNames = []string{"cert_hostname", "alert_email"}
	notAfter   = NewGauge("cert_not_after", "cert not after X Unix Epoch seconds ", labelNames)
	config     Config
	hostQueue  = make(chan Host, 10)
)

func init() {
	logger, _ = zap.NewProduction()
}

type Host struct {
	Address     string   `toml:"address"`
	AlertEmails []string `toml:"alert-emails"`
}
type Config struct {
	ListenAddress         string `toml:"listen-address"`
	RefreshIntervalSecond int    `toml:"refresh-interval-second"`
	Concurrency           int    `toml:"concurrency"`
	Hosts                 []Host `toml:"hosts"`
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

	// 等待程序退出
	go func(ctx context.Context) {
		<-signalCh
		cancel()
		httpServer.Shutdown(ctx)
	}(ctx)

	// 收集所有hosts
	go runCollectHosts(ctx, time.Duration(config.RefreshIntervalSecond)*time.Second)

	// 检查过期时间
	go processHosts(ctx)

	logger.Info("http server started", zap.String("address", config.ListenAddress),
		zap.Int("ssl check concurrency", config.Concurrency))
	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		logger.Fatal("HTTP server ListenAndServe Error: %v", zap.Error(err))
	}

	logger.Info("Bye")
}

func parseConfig() Config {
	var c Config
	_, err := toml.DecodeFile(*configFile, &c)
	if err != nil {
		logger.Error("无法Decode toml 配置文件", zap.Error(err))
		os.Exit(1)
	}
	return c
}

func collectHosts(ctx context.Context) {
	c := parseConfig()

	for _, host := range c.Hosts {
		hostQueue <- host
		logger.Info("collectHosts", zap.String("address", host.Address),
			zap.Strings("emails", host.AlertEmails))
	}
}

func runCollectHosts(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 程序启动就先执行一次，然后再 Ticker 周期性运行
	go collectHosts(ctx)

	for {
		select {
		case <-ticker.C:
			go collectHosts(ctx)
		case <-ctx.Done():
			logger.Info("collectHosts ctx done")
			return
		}
	}
}
