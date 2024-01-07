# ssl-certs-check

Simple SSL Certs Expiration Check

#### Features

- config the **hosts** and **alert emails** inside toml configuration file
- docker-compose start **Prometheus/Alertmanager/Grafana** for check and alert

#### How it works

- hosts ssl certs will be checked regulaly by ssl-certs-check,
- expose expiration date as prometheus metrics
- base on configuration, all metrics have alert email as labels
- generated alertmanager config file base on configuration for alert

## Building Binary

    make build
    cp configurations/config-example.toml configurations/config.toml
    # modify configurations/config.toml, then
    ./ssl-certs-check -config configurations/config.toml

## Docker build

modify `docker-compose.yaml` ssl-certs-check env `ENV_GOPROXY`, then

    docker-compose build

## Configuration

### ssl-certs-check main config file: [configurations/config-example.toml](configurations/config-example.toml)

- **smtp-xxxx and [[hosts]]** related configuration need to be modified

```toml
    listen-address = ":8080"

    # refresh to get latest hosts 
    refresh-interval-second=3600

    [alertmanager]
    # after hosts change, ssl-certs-check will call this url to reload alertmanager
    reload-url="http://alertmanager:9093/-/reload"

    # ssl-certs-check will generate alertmanager.conf to this path
    config-path="configurations/alertmanager.conf"

    # altermanager will use these smtp server send alert emails
    smtp-smarthost=''
    smtp-from=''
    smtp-username=''
    smtp-password=''


    # hosts example: 
    # - if port not provided, default is 443
    # - alert-emails define who care about this address' cert expiration

    [[hosts]]
        address = "www.supertechfans.com"
        alert-emails = ["u1@example.com", "u2@example.com"]
    [[hosts]]
        address = "githube.com:443"
        alert-emails = ["abc@example.com"]
```

### alert rule [configurations/alert_rules.yml](configurations/alert_rules.yml)

- You can adjust the alert expiration days (25 here)

```yaml
    groups:
  - name: 'ssl-certs-check-alert'
        rules:
    - alert: SSLCertsNearlyExpiration
            expr: round((exporter_cert_not_after{} - time())/3600/24) < 25
            annotations:
            title: 'SSL Certs Will expire after {{ $value }} days'
            description: ' Please kindly renew'
            labels:
            severity: 'critical'
```

## Usage

    docker-compose up -d

Then access:

- [alertmanager](http://localhost:9093/)
- [prometheus](http://localhost:9090/graph?g0.expr=round((exporter_cert_not_after%20-%20time())%20%2F%203600%20%2F%2024)&g0.tab=1&g0.stacked=0&g0.show_exemplars=0&g0.range_input=1h)
- [grafana](http://localhost:3000/) (admin/admin)

## Metrics

| Metric                     | Meaning                                                        | Labels                    |
| -------------------------- | -------------------------------------------------------------- | ------------------------- |
| exporter_cert_not_after    | cert not after X Unix Epoch seconds                            | cert_hostname,alert_email |
| exporter_host_queue_length | how many hosts in queue waiting to be check (lower the better) |                           |
