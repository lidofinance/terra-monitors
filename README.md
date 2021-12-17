# terra-monitors

Before running the service you have to fill .env file `./docker/env/.telegram.env` with the following variables:
```shell
cat ./docker/env/.telegram.env 
TELEGRAM_BOTTOKEN=<telegram_bottoken>
TELEGRAM_CHAT_ID=<telegram_chat_id>
```

Variables required to service work.\
You should pass the .env file and pass it to docker-compose service with `--env-file` argument.
```shell
# custom endpoint for alert messages 
GF_SERVER_ROOT_URL=http://swelf-host:3000

# Grafana frontend port - optional value, default value is 3000
GRAFANA_PORT=3000 

# LCD - light client daemon, https://docs.terra.money/terracli/lcd.html
LCD_ENDPOINTS=fcd.terra.dev
LCD_SCHEMES=https
# terra-monitor data update interval
UPDATE_DATA_INTERVAL=30s

# monitored contracts
ADDRESSES_HUB_CONTRACT=terra1mtwph2juhj0rvjz7dy92gvl6xvukaxu8rfv8ts
ADDRESSES_REWARD_CONTRACT=terra17yap3mhph35pcwvhza38c2lkj7gzywzy05h7l0
ADDRESSES_BLUNA_TOKEN_INFO_CONTRACT=terra1kc87mu460fwkqte29rquh4hc20m54fxwtsx7gp
ADDRESSES_VALIDATORS_REGISTRY_CONTRACT=terra_dummy_validators_registry
ADDRESSES_REWARDS_DISPATCHER_CONTRACT=terra_dummy_rewards_dispatcher
ADDRESSES_AIR_DROP_REGISTRY_CONTRACT=terra_dummy_airdrop

# monitored bot, executing update_global_index message on the hub contract
# https://www.notion.so/bAsset-index-updating-bot-f64ebb5ec6704f05a840d93f28b1e3be
ADDRESSES_UPDATE_GLOBAL_INDEX_BOT_ADDRESS=terra1eqpx4zr2vm9jwu2vas5rh6704f6zzglsayf2fy

# Version of basset contracts monitored, default value is 2
BASSET_CONTRACTS_VERSION=1

# service name is a tag value to detect log entries related to the our service instance only
# default value is lido_terra
# should be set explicitly to uniq values in case multiple monitoring instance works on same machine
SERVICE_NAME=lido_terra_mainnet

# This configuration controls the number of mean absolute deviations
# (https://en.wikipedia.org/wiki/Median_absolute_deviation) that a validator's total delegations
# amount should be grater than the median delegations amount. Higher values mean less monitor
# sensitivity. A step of 0.25 is nice for calibration.
DELEGATIONS_DISTRIBUTION_CONFIG_NUM_MEDIAN_ABSOLUTE_DEVIATIONS=3
```

**N.B.: you can specify failover endpoints (sorted by priority, max to min) for the `LCD_ENDPOINTS` config:**

```
LCD_ENDPOINTS=fcd.terra.dev,scp.terra.dev
```

To run the service with env file - `./docker/env/.lido_terra.env`, `./docker/env/.lido_terra.env` is not being tracked by a git, and could be changed for any purpose.
```shell
make start
```

To stop the service:
```shell
make stop
```

To run with predefined testnet `./docker/env/.lido_terra.testnet.env` monitoring\
port = 3000
```shell
make start_testnet
```

To run with predefined mainnet `./docker/env/.lido_terra.prod.env`monitoring\
port = 3001
```shell
make start_mainnet
```

These two commands start_testnet and start_mainnet can be run from same 
directory due to cli docker-compose parameter - `-p terra_monitors_mainnet` 
without mixing up data

Grafana dashboards available at http://127.0.0.1:3000.

Default login/pass: `admin/admin`.
