# ssl-certs-check

Simple SSL Certs Expiration Check

- config the hosts and alert email inside toml configuration file
- programe will check these hosts ssl certs regulaly
- then expose the ssl certs expiration date as prometheus metrics
- all metrics with alert email as label
- auto generate alertmanager config file base on configuration
- docker-compose will start Prometheus/Alertmanager/Grafana for check and alert

## Building Binary

    make build
    ./ssl-certs-check -config configurations/config.toml

### Docker build

modify `docker-compose.yaml` ssl-certs-check env `ENV_GOPROXY`, then

    docker-compose build

## Usage

    docker-compose up

Then access:

- [alertmanager](http://localhost:9093/)
- [prometheus](http://localhost:9090/)
- [grafana](http://localhost:3000/)

## Metrics

| Metric                     | Meaning                                                        | Labels                    |
| -------------------------- | -------------------------------------------------------------- | ------------------------- |
| exporter_cert_not_after    | cert not after X Unix Epoch seconds                            | cert_hostname,alert_email |
| exporter_host_queue_length | how many hosts in queue waiting to be check (lower the better) |                           |
