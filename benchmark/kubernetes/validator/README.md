# Validator Helm Chart

A validator is an Ethereum client to validate transactions and generate blocks.

## Prerequisites
* [StatefulSets](https://kubernetes.io/docs/concepts/abstractions/controllers/statefulsets/) support
* Dynamically provisioned persistent volumes

## Installing the Chart

To install the chart:

```bash
helm install \
  --name ${DEPLOYMENT_ENVIRONMENT}-transactor-${NAME} \
  kubernetes/amis/generic/validator
```

## Configuration

The following tables lists the configurable parameters of the transactor chart and their default values.

| Parameter                       | Description                                       | Default                                                    |
| ------------------------------- | ------------------------------------------------- | ---------------------------------------------------------- |
| `service.type`                  | The Kubernetes service type                       | `ClusterIP`                                                |
| `service.externalPeerPort`      | Exposed Ethereum P2P port                         | `30303`                                                    |
| `service.ExternalRPCPort`       | Exposed Ethereum RPC port                         | `8545`                                                     |
| `service.ExternalWSPort`        | Exposed Ethereum WebSocket port                   | `8546`                                                     |
| `replicaCount`                  | Kubernetes statefulset replicas                   | `1`                                                        |
| `image.repository`              | Docker image repository                           | `ethereum/client-go`                                       |
| `image.tag`                     | Docker image tag                                  | `alpine`                                                   |
| `image.pullPolicy`              | Docker image pulling policy                       | `Always`                                                   |
| `ethereum.identity`             | Custom node name                                  | `$POD_NAME`                                                |
| `ethereum.port`                 | Network listening port                            | `30303`                                                    |
| `ethereum.networkID`            | Network identifier                                | `12345`                                                    |
| `ethereum.cache`                | Megabytes of memory allocated to internal caching | `512`                                                      |
| `ethereum.rpc.enabled`          | Enable the HTTP-RPC server                        | `false`                                                    |
| `ethereum.rpc.addr`             | HTTP-RPC server listening interface               | `localhost`                                                |
| `ethereum.rpc.port`             | HTTP-RPC server listening port                    | `8545`                                                     |
| `ethereum.rpc.api`              | API's offered over the HTTP-RPC interface         | `"eth,net,web3,personal"`                                  |
| `ethereum.rpc.corsdomain`       | Comma separated list of domains from which to accept cross origin requests | `*`                               |
| `ethereum.ws.enabled`           | Enable the WS-RPC server                          | `false`                                                    |
| `ethereum.ws.addr`              | WS-RPC server listening interface                 | `localhost`                                                |
| `ethereum.ws.port`              | WS-RPC server listening port                      | `8546`                                                     |
| `ethereum.ws.api`               | API's offered over the WS-RPC interface           | `"eth,net,web3,personal"`                                  |
| `ethereum.ws.origins`           | Origins from which to accept websockets requests  | `*`                                                        |
| `ethereum.mining.enabled`       | Enable mining                                     | `true`                                                     |
| `ethereum.mining.threads`       | Number of CPU threads to use for mining           | `2`                                                        |
| `ethereum.mining.etherbase`     | Public address for block mining rewards           | `"1a9afb711302c5f83b5902843d1c007a1a137632"`               |
| `ethereum.ethstats.enabled`     | Enable ethstats reporting                         | `true`                                                     |
| `ethereum.ethstats.addr`        | Ethstats service address                          | `ws://eth-netstats`                                        |
| `ethereum.ethstats.port`        | Ethstats service port                             | `3000`                                                     |
| `ethereum.ethstats.secret`      | Ethstats service websocket secret                 | `bb98a0b6442386d0cdf8a31b267892c1`                         |
| `ethereum.nodekey.hex`          | P2P node key as hex                               | `aaaaaaaaaaaaaa5302ccdd5b6ffa092e148ba497e352c2824f366ec399072068` |
| `ethereum.account.address`      | Account address assigned to the validator         | `1a9afb711302c5f83b5902843d1c007a1a137632`                         |
| `ethereum.account.data`         | Full account data file as JSON string             | `{"address":"1a9afb711302c5f83b5902843d1c007a1a137632","Crypto":{"cipher":"aes-128-ctr","ciphertext":"132b50d7c8944a115824de7c00911c40a90f84f27c614b7a3ef05ee8fd414312","cipherparams":{"iv":"0f745599d1b3303988ce210fb82b8c7f"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"bce940bac232b4a9c5a2d50e5be51fde5cecfa7da9d49d8f650f91167bebf0de"},"mac":"36d515070b797aec58a574a3e04ea109498ee7674b15d7f952322cda7dcb68e3"},"id":"5d212b4c-3dd0-4c52-a32f-e42bf1b41133","version":3}`                         |
| `volumes.chaindir.size`         | The maximum size usage of Ethereum data directory | `10Gi`                                                     |
| `volumes.chaindir.storageClass` | The Kubernetes storage class of Ethereum data     | `$NAMESPACE-chaindata`                                     |
| `global.computingSpec`          | The computing spec level of node to schedule      | `low`                                                      |
| `global.nodeType`               | The type of node to schedule                      | `generic`                                                  |


Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```bash
helm install \
  --name ${DEPLOYMENT_ENVIRONMENT}-validator-${NAME} \
  -f values.yaml \
  kubernetes/amis/generic/validator
```

> **Tip**: You can use the default [values.yaml](values.yaml)