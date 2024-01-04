# ssl-certs-check

Simple SSL Certs Check

- config the hosts and alert email inside toml configuration file
- programe will check these hosts ssl certs regulaly
- then expose the ssl certs expiration date as prometheus metrics
- all metrics with alert email as label
