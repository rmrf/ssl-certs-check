package main

import (
	"context"
	"crypto/tls"
	"math"
	"net"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

type certErrors struct {
	commonName string
	errs       []error
}

type hostResult struct {
	address string
	err     error
	certs   []certErrors
}

func processHosts(ctx context.Context) {

	results := make(chan hostResult)

	var wg sync.WaitGroup
	wg.Add(config.Concurrency)
	for i := 0; i < config.Concurrency; i++ {
		go func() {
			processQueue(ctx, hostQueue, results)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		if r.err != nil {
			logger.Warn("cert err", zap.Error(r.err), zap.String("address", r.address))
			continue
		}
		for _, cert := range r.certs {
			for _, err := range cert.errs {
				logger.Warn("cert err", zap.Error(err), zap.String("address", r.address))
			}
		}
	}

}

func processQueue(ctx context.Context, hosts <-chan Host, results chan<- hostResult) {

	hostQueueLen.WithLabelValues(config.ListenAddress).Set(float64(len(hostQueue)))
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	for host := range hosts {
		select {
		case results <- checkHost(host):
		case <-ticker.C:
			hostQueueLen.WithLabelValues(config.ListenAddress).Set(float64(len(hostQueue)))
		case <-ctx.Done():
			logger.Info("proccessQueue ctx done")
			return
		}
	}
}

func checkHost(host Host) (result hostResult) {
	logger.Info("checkhost", zap.String("address", host.Address))
	result = hostResult{
		address: host.Address,
		certs:   []certErrors{},
	}
	var address = host.Address
	if !strings.Contains(address, ":") {
		address += ":443"
	}

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 10 * time.Second}, "tcp", address, nil)
	if err != nil {
		result.err = err
		return
	}
	defer conn.Close()
	var notAfterUnix = math.MaxInt64

	checkedCerts := make(map[string]struct{})
	for _, chain := range conn.ConnectionState().VerifiedChains {
		for _, cert := range chain {
			if _, checked := checkedCerts[string(cert.Signature)]; checked {
				continue
			}
			checkedCerts[string(cert.Signature)] = struct{}{}
			cErrs := []error{}

			// Check the expiration, find out the shortest expiration in the chain
			if !cert.NotAfter.IsZero() && int(cert.NotAfter.Unix()) < notAfterUnix {
				notAfterUnix = int(cert.NotAfter.Unix())
			}

			result.certs = append(result.certs, certErrors{
				commonName: cert.Subject.CommonName,
				errs:       cErrs,
			})
		}
	}

	for _, e := range host.AlertEmails {
		notAfter.WithLabelValues(address, e).Set(float64(notAfterUnix))
	}

	return
}
