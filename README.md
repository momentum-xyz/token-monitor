# token-service

The token service monitors accounts for token transfers to allow for token gating. Its initial purpose is to enable Token Gated Access for Odyssey Momentum. It receives a list of rules and a list of web3
addresses and starts a listener that listens for logs on the specified networks in the rules structure.

The fetched logs are then scanned for the web3 addresses initially provided. The backend is notified if access for a
user needs to be updated corresponding to a particular rule.

This repository contains a config.dev.yaml with an MQTT configuration and a list of networks.


## Pre-requisites:
- Go >=1.17.2
- Make
- Docker

## Getting started

1. Run `make up`
2. Run `make run`

## Deployment

Build the token service with `make build`.

Configure the config path with the `CONFIG_PATH` environment variable. If no environment variable is found the token_service loads its configuration from `config.dev.yaml`, which can serve as a template for custom configurations.

## The caching layer
The token service uses a cache to store user balances. The initialization of the token service and collecting the initial token balances of a list of users requires a large number of requests, that otherwise would take a long time to accumulate. Therefore, the cache stores user balances to prevent a long initialization time. There are two caching mechanisms:

In order to determine which caching implementation is used the config field `cache.type` needs to be filled with either "bbolt" or "redis".
### Using bbolt
Bbolt is a key-value database that stores its data on the filesystem. It requires a path to your db file, a filemode and a bucket name. The filemode determines the read and write restrictions to the db file using the format known from unix file permissions (default = 0600). Configuration fields can be found in `pkg/cache/cache.go`.

! IMPORTANT !
Only one instance of the token service can concurrently access the db file. Therefore, using bbolt limits the horizontal scaling option of the token service.

### Using redis
Redis can be used as a key-value cache as well. Configuration fields can be found in `pkg/cache/cache.go`.

## Token Service MQTT Interface

The service expects messages as JSON strings on the following topics:

### Receiving Topic: token-service/active-users

``` 
[
    {
        "ethereum_address": "0x0"
    }
]
```

### Receiving Topic: token-service/active-rules

```
[
    {
        "id":number,
        "active": bool,
        "token": {
            "type": string,
            "address": string,
            "token_id": number
        },
        "network": string,
        "requirements": {
            "minimumBalance": number
        }
    }
]
```

### Receiving Topic: token-service/rules

```
{
    "id":number,
    "active": bool,
    "token": {
        "type": string,
        "address": string,
        "token_id": number
    },
    "network": string,
    "requirements": {
        "minimum_balance": number
    }
}
```

### Receiving Topic: token-service/user-event

```
{
    accountAddress: string;
}
```

The token service produces messages on the `permission_updates` topic.

### Publishing Topic: permission_updates

```
{
    accountAddress: string;
    ruleUUID: string;
    active: boolean;
}
```

## Contributors ✨

The token service is a project initiated by Odyssey. Thanks to these contributors 😎

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>

  <tr>
  <td align="center"><a href="https://github.com/jellevdp"><img src="https://avatars.githubusercontent.com/jellevdp?v=3?s=100" width="100px;" alt=""/><br /><sub><b>Jelle van der Ploeg </b></sub></a><br />
    </td>
<td align="center"><a href="https://github.com/tech-sam"><img src="https://avatars.githubusercontent.com/tech-sam?v=3?s=100" width="100px;" alt=""/><br /><sub><b>Sumit</b></sub></a><br />
</td>
   <td align="center"><a href="https://github.com/e-nikolov"><img src="https://avatars.githubusercontent.com/e-nikolov" width="100px;" alt=""/><br /><sub><b>Emil Nikolov  </b></sub></a><br />
    </td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/nwasiqUC"><img src="https://avatars.githubusercontent.com/nwasiqUC" width="100px;" alt=""/><br /><sub><b>Wasiq  </b></sub></a><br />
    </td>
    <td align="center"><a href="https://github.com/antst"><img src="https://avatars.githubusercontent.com/antst" width="100px;" alt=""/><br /><sub><b>Anton Starikov</b></sub></a><br />
    </td>
    <td align="center"><a href="https://github.com/jor-rit"><img src="https://avatars.githubusercontent.com/jor-rit" width="100px;" alt=""/><br /><sub><b>Jorrit</b></sub></a><br />
    </td>
  </tr>
</table>