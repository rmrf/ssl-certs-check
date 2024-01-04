#!/bin/bash


curl $APOLLO_META/configs/3734/default/application > /lain/app/application.json
cat /lain/app/application.json | jq -r '.configurations["ssl-certs-check.toml"]' > /lain/app/ssl-certs-check.toml

exec /lain/app/ssl-certs-check -config /lain/app/ssl-certs-check.toml

