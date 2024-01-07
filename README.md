# ssl-certs-check

Simple SSL Certs Expiration Check

Features:

- config the **hosts** and **alert emails** inside toml configuration file
- docker-compose start Prometheus/Alertmanager/Grafana for check and alert

How it works:

- hosts ssl certs will be checked regulaly by ssl-certs-check,
- expose expiration date as prometheus metrics
- base on configuration, all metrics have alert email as labels
- generated alertmanager config file base on configuration for alert

## Building Binary

    make build
    cp configurations/config-example.toml configurations/config.toml
    # modify configurations/config.toml, then
    ./ssl-certs-check -config configurations/config.toml

### Docker build

modify `docker-compose.yaml` ssl-certs-check env `ENV_GOPROXY`, then

    docker-compose build

## Usage

    docker-compose up -d

Then access:

- [alertmanager](http://localhost:9093/)
- [prometheus](http://localhost:9090/)
- [grafana](http://localhost:3000/) (admin/admin)

## Metrics

| Metric                     | Meaning                                                        | Labels                    |
| -------------------------- | -------------------------------------------------------------- | ------------------------- |
| exporter_cert_not_after    | cert not after X Unix Epoch seconds                            | cert_hostname,alert_email |
| exporter_host_queue_length | how many hosts in queue waiting to be check (lower the better) |                           |
