# istanbul-tools

[![Build Status](https://travis-ci.com/jpmorganchase/istanbul-tools.svg?branch=master)](https://travis-ci.com/jpmorganchase/istanbul-tools) [![Download](https://api.bintray.com/packages/quorumengineering/istanbul-tools/istanbul/images/download.svg)](https://bintray.com/quorumengineering/istanbul-tools/istanbul/_latestVersion)

`istanbul-tools` contains tools for configuring Istanbul BFT (IBFT) and QBFT networks, integration tests for both IBFT Geth and Quorum, and load testing utilities for IBFT Geth.

## Build istanbul command line interface
* Go 1.15+
```
$ make istanbul
$ ./build/bin/istanbul --help

NAME:
   istanbul - the istanbul-tools command line interface

USAGE:
   istanbul [global options] command [command options] [arguments...]

VERSION:
   v1.1.0

COMMANDS:
     extra    Istanbul extraData manipulation
     setup    Setup your Istanbul network in seconds
     reinit   Reinitialize a genesis block using previous node info
     address  Extract validator address
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version

COPYRIGHT:
   Copyright 2017 The AMIS Authors                        
```

### `extra` subcommand

<details>
<summary>Click here to expand</summary>

`extra` helps generate RLP-encoded extra data in `ExtraData` field of the genesis block. Extra data is composed of signer vanity and `IstanbulExtra`. `IstanbulExtra` is defined as follows:

```go
type IstanbulExtra struct {
    Validators    []common.Address  // Validator addresses
    Seal          []byte            // Proposer seal 65 bytes
    CommittedSeal [][]byte          // Committed seal, 65 * len(Validators) bytes
}
```

**Note**: `Seal` and `CommittedSeal` are not considered in genesis block.

```sh
$ ./build/bin/istanbul extra

NAME:
   istanbul extra - Istanbul extraData manipulation

USAGE:
   istanbul extra command [command options] [arguments...]

COMMANDS:
     decode  To decode an Istanbul extraData
     encode  To encode an Istanbul extraData

OPTIONS:
   --help, -h  show help

```

#### `extra` examples

##### `encode` subcommand

Encode the given file to extra data.

```
$ ./build/bin/istanbul extra encode --config ./cmd/istanbul/example/config.toml

OUTPUT:
Encoded Istanbul extra-data: 0x0000000000000000000000000000000000000000000000000000000000000000f89af85494475cc98b5521ab2a1335683e7567c8048bfe79ed9407d8299de61faed3686ba4c4e6c3b9083d7e2371944fe035ce99af680d89e2c4d73aca01dbfc1bd2fd94dc421209441a754f79c4a4ecd2b49c935aad0312b8410000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0
```

##### `decode` subcommand

Decode extra data from the given input.

```
$ ./build/bin/istanbul extra decode --extradata 0x0000000000000000000000000000000000000000000000000000000000000000f89af85494475cc98b5521ab2a1335683e7567c8048bfe79ed9407d8299de61faed3686ba4c4e6c3b9083d7e2371944fe035ce99af680d89e2c4d73aca01dbfc1bd2fd94dc421209441a754f79c4a4ecd2b49c935aad0312b8410000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0

OUTPUT:
vanity:  0x0000000000000000000000000000000000000000000000000000000000000000
validator:  0x475cc98b5521ab2a1335683e7567c8048bfe79ed
validator:  0x07d8299de61faed3686ba4c4e6c3b9083d7e2371
validator:  0x4fe035ce99af680d89e2c4d73aca01dbfc1bd2fd
validator:  0xdc421209441a754f79c4a4ecd2b49c935aad0312
seal: 0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
```

</details>

### `setup` subcommand

<details>
<summary>Click here to expand</summary>

When `--nodes --verbose` flags are given, a `static-nodes.json` template as well as the validators' node keys, public keys, and addresses are generated. When `--docker-compose` is given, a `docker-compose.yml` for the validators is generated. When `--save` flag is given, all generated configs will be saved. Use Quorum when `--quorum` flag is given.

**Note**: the generated `static-nodes.json` file has the IP and port values defined using the flags `--nodeIp`, `--nodePortBase`, and `--nodePortIncrement`.
 If these flags are not provided then the IP `0.0.0.0` and port `30303` will be used; in which case, the `static-nodes.json` will require manual changes to match your environment.

#### `setup` examples

```
$ ./build/bin/istanbul setup --num 4 --nodes --verbose --nodeIp 127.0.0.1 --nodePortBase 21000 --nodePortIncrement 1
validators
{
    "Address": "0x5e5d0e2b80005a7e1f93044ddd64b2df0f8e488d",
    "Nodekey": "e5f9b868651ea8f4883744f2753ead9dfcdf7b1d8a96de0e733f406938dca1eb",
    "NodeInfo": "enode://8759a8a6921be78ec4e66ec77ae26ba9b3b1a51d1f83b16683c6f25e5a1d95a4de2c5bf4c2c05e1b984fae440236d96063efe933425df72659ee9de824cda6e1@127.0.0.1:21000?discport=0"
}
{
    "Address": "0x1b706dd850229813ee7c4002cd2fedc91380bb5a",
    "Nodekey": "2c13ee666b2ce617bf1e0d7fe7c8f058be27ea3a1aaabbfc63570a65f0bdae38",
    "NodeInfo": "enode://40dd1e7ba45e5bcd242420986d9d03133ce49399c6197e43254d523e94f547532d4c47c8aaba4b000c5a718568a48013b035c86f3ed8b13248888a15a76761c1@127.0.0.1:21001?discport=0"
}
{
    "Address": "0xdfdf27987b042bb3706d3a7c4b60e80a645744de",
    "Nodekey": "8bbf54eace8738f9d3ee90d5b949951f43d89acdb4b883d9188a141bdcd0153e",
    "NodeInfo": "enode://d188378b3eef56584b8ebd3da3ad579d39d23511943573cdeae5b8a37b5df22c369bf8900c4f42a9d4d5e55bc3cd357f319de8f833db3232295be22c8accc006@127.0.0.1:21002?discport=0"
}
{
    "Address": "0x5950b8f849daf1a78e119648c79111721353df59",
    "Nodekey": "9179c038483a2547c39f77f121065231d84a9c8d9bd044e1ddc19f653a23c751",
    "NodeInfo": "enode://d855be48593e6f2dd6201334e9381a2f01dac4a847385a393b1f664503b7b7020326e9f3f84f2d5713bf360d16566ed2b84d7df0b8b8313a7a4c4cf087ccfe27@127.0.0.1:21003?discport=0"
}



static-nodes.json
[
    "enode://8759a8a6921be78ec4e66ec77ae26ba9b3b1a51d1f83b16683c6f25e5a1d95a4de2c5bf4c2c05e1b984fae440236d96063efe933425df72659ee9de824cda6e1@127.0.0.1:21000?discport=0",
    "enode://40dd1e7ba45e5bcd242420986d9d03133ce49399c6197e43254d523e94f547532d4c47c8aaba4b000c5a718568a48013b035c86f3ed8b13248888a15a76761c1@127.0.0.1:21001?discport=0",
    "enode://d188378b3eef56584b8ebd3da3ad579d39d23511943573cdeae5b8a37b5df22c369bf8900c4f42a9d4d5e55bc3cd357f319de8f833db3232295be22c8accc006@127.0.0.1:21002?discport=0",
    "enode://d855be48593e6f2dd6201334e9381a2f01dac4a847385a393b1f664503b7b7020326e9f3f84f2d5713bf360d16566ed2b84d7df0b8b8313a7a4c4cf087ccfe27@127.0.0.1:21003?discport=0"
]



genesis.json
{
    "config": {
        "chainId": 2017,
        "homesteadBlock": 0,
        "eip150Block": 0,
        "eip150Hash": "0x0000000000000000000000000000000000000000000000000000000000000000",
        "eip155Block": 0,
        "eip158Block": 0,
        "byzantiumBlock": 0,
        "constantinopleBlock": 0,
        "istanbul": {
            "epoch": 30000,
            "policy": 0
        },
        "isQuorum": true,
        "txnSizeLimit": 64
    },
    "nonce": "0x0",
    "timestamp": "0x5a093aac",
    "extraData": "0x0000000000000000000000000000000000000000000000000000000000000000f89af854945e5d0e2b80005a7e1f93044ddd64b2df0f8e488d941b706dd850229813ee7c4002cd2fedc91380bb5a94dfdf27987b042bb3706d3a7c4b60e80a645744de945950b8f849daf1a78e119648c79111721353df59b8410000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0",
    "gasLimit": "0x47b760",
    "difficulty": "0x1",
    "mixHash": "0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365",
    "coinbase": "0x0000000000000000000000000000000000000000",
    "alloc": {
        "1b706dd850229813ee7c4002cd2fedc91380bb5a": {
            "balance": "0x446c3b15f9926687d2c40534fdb564000000000000"
        },
        "5950b8f849daf1a78e119648c79111721353df59": {
            "balance": "0x446c3b15f9926687d2c40534fdb564000000000000"
        },
        "5e5d0e2b80005a7e1f93044ddd64b2df0f8e488d": {
            "balance": "0x446c3b15f9926687d2c40534fdb564000000000000"
        },
        "dfdf27987b042bb3706d3a7c4b60e80a645744de": {
            "balance": "0x446c3b15f9926687d2c40534fdb564000000000000"
        }
    },
    "number": "0x0",
    "gasUsed": "0x0",
    "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000"
}
```

```
$ ./build/bin/istanbul setup --help
NAME:
   istanbul setup - Setup your Istanbul network in seconds

USAGE:
   istanbul setup [command options] [arguments...]

DESCRIPTION:
   This tool helps generate:

    * Genesis block
    * Static nodes for all validators
    * Validator details

    for Istanbul consensus.


OPTIONS:
   --num value       Number of validators (default: 0)
   --nodes           Print static nodes template
   --verbose         Print validator details
   --quorum          Use quorum
   --docker-compose  Print docker compose file
   --save            Save to files
```

</details>

### `address` subcommand

<details>
<summary>Click here to expand</summary>

This command is to extract Validator Address (ID) from node key hex which is the node private key in hex

E.g.: 
```
$ ./build/bin/istanbul address --nodekeyhex 1be3b50b31734be48452c29d714941ba165ef0cbf3ccea8ca16c45e3d8d45fb0
0xd8dba507e85f116b1f7e231ca8525fc9008a6966
```
</details>

## Build qbft command line interface
* Go 1.15+
```
$ make qbft
$ ./build/bin/qbft --help

NAME:
   qbft - the qbft command line interface

USAGE:
   qbft [global options] command [command options] [arguments...]

COMMANDS:
   extra    qbft extraData manipulation
   setup    Setup your qbft network in seconds
   reinit   Reinitialize a genesis block using previous node info
   address  Extract validator address
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version

COPYRIGHT:
   Copyright 2017 The AMIS Authors                      
```

### `extra` subcommand

<details>
<summary>Click here to expand</summary>

`extra` helps generate RLP-encoded extra data in `ExtraData` field of the genesis block. Extra data is composed of `QBFTExtra`. `QBFTExtra` is defined as follows:

```go
type QBFTExtra struct {
VanityData    []byte
Validators    []common.Address
Vote          *ValidatorVote
Round         uint32
CommittedSeal [][]byte
}
```

**Note**: `VanityData`, `Vote`, `Round` and `CommittedSeal` are not considered in genesis block.

```sh
$ ./build/bin/qbft extra

NAME:
   qbft extra - qbft extraData manipulation

USAGE:
   qbft extra command [command options] [arguments...]

COMMANDS:
   decode  To decode an qbft extraData
   encode  To encode an qbft extraData

OPTIONS:
   --help, -h  show help

```

#### `extra` examples

##### `encode` subcommand

Encode the given file to extra data.

```
$ ./build/bin/qbft extra encode --config ./cmd/istanbul/example/config.toml

OUTPUT:
Encoded qbft extra-data: 0xf87aa00000000000000000000000000000000000000000000000000000000000000000f85494475cc98b5521ab2a1335683e7567c8048bfe79ed9407d8299de61faed3686ba4c4e6c3b9083d7e2371944fe035ce99af680d89e2c4d73aca01dbfc1bd2fd94dc421209441a754f79c4a4ecd2b49c935aad0312c080c0
```

##### `decode` subcommand

Decode extra data from the given input.

```
$ ./build/bin/qbft extra decode --extradata 0xf87aa00000000000000000000000000000000000000000000000000000000000000000f85494475cc98b5521ab2a1335683e7567c8048bfe79ed9407d8299de61faed3686ba4c4e6c3b9083d7e2371944fe035ce99af680d89e2c4d73aca01dbfc1bd2fd94dc421209441a754f79c4a4ecd2b49c935aad0312c080c0

OUTPUT:
vanity:  0x0000000000000000000000000000000000000000000000000000000000000000
validator:  0x475cc98B5521AB2A1335683e7567c8048BfE79eD
validator:  0x07D8299de61FAeD3686BA4c4e6c3B9083d7e2371
validator:  0x4fe035CE99AF680d89e2c4D73aCA01DBFc1Bd2FD
validator:  0xdC421209441A754F79C4A4eCD2b49c935AAD0312
round:  0
```

</details>

### `setup` subcommand

<details>
<summary>Click here to expand</summary>

When `--nodes --verbose` flags are given, a `static-nodes.json` template as well as the validators' node keys, public keys, and addresses are generated. When `--save` flag is given, all generated configs will be saved. Use Quorum when `--quorum` flag is given.

**Note**: the generated `static-nodes.json` file has the IP and port values defined using the flags `--nodeIp`, `--nodePortBase`, and `--nodePortIncrement`.
If these flags are not provided then the IP `0.0.0.0` and port `30303` will be used; in which case, the `static-nodes.json` will require manual changes to match your environment.

#### `setup` examples

```
$ ./build/bin/qbft setup --num 4 --nodes --verbose --nodeIp 127.0.0.1 --nodePortBase 21000 --nodePortIncrement 1
validators
{
	"Address": "0x82f114cde4898983626a27af4eb928ff804e60ee",
	"Nodekey": "34a43df6fc32c3c8561024fd0fbd744eebad5cebb74bc1c3ffa4a8bb2136489b",
	"NodeInfo": "enode://b4a8371baf676ec384ec6e97c33e30e33e7d68460dc9014a8262258f933eb6476e433017e9f1f56e7fc5b2eef9a5655a6426cb491133f64ce1c7f9d08e0c6fa4@127.0.0.1:21000?discport=0"
}
{
	"Address": "0x21de2bf49c07595cf8c7c64ac5b173a112171cfe",
	"Nodekey": "8a3f9c2ff1b17374d521da6be6a33a77722d84d223ccf41ee9c8e7e9107e1944",
	"NodeInfo": "enode://be355d2f3e884be7f5a1698e105de2408ebb921632dfe608a95e5eddfe4352c70d8f21f8f2ecf59658749b97bf1bdfb939c3d58532f1e511f22f827732ef6043@127.0.0.1:21001?discport=0"
}
{
	"Address": "0x3f3e6c684f16f7fac8aa6ae20bb04a3910367994",
	"Nodekey": "ffe595c61270e51a9f51e6a5288bb9f19e930a264d11243e865508104ae6498c",
	"NodeInfo": "enode://6f6c58bb00418b6892a59936717f597eb8a2e5113a5e68188bb958275f1c5f767fae2db4e9f675970c8b20d9f4da2a8e0066292725756446348eb6ea10ac6215@127.0.0.1:21002?discport=0"
}
{
	"Address": "0xe63a320b26610685d3d5124c7c65b360735ab8f2",
	"Nodekey": "321339fbbc84f71c7b0c8aebe2bf53a64951a3bfee75960edaa4313b614f4a3c",
	"NodeInfo": "enode://70465abd32f3bd107eb6cce1b99c57db154b98039efecfcb757b1b8244cbb26846025fe8a1eadfeaff9a7f49d624d86f82ade87b737233588f36fce78f3fa7b9@127.0.0.1:21003?discport=0"
}



static-nodes.json
[
	"enode://b4a8371baf676ec384ec6e97c33e30e33e7d68460dc9014a8262258f933eb6476e433017e9f1f56e7fc5b2eef9a5655a6426cb491133f64ce1c7f9d08e0c6fa4@127.0.0.1:21000?discport=0",
	"enode://be355d2f3e884be7f5a1698e105de2408ebb921632dfe608a95e5eddfe4352c70d8f21f8f2ecf59658749b97bf1bdfb939c3d58532f1e511f22f827732ef6043@127.0.0.1:21001?discport=0",
	"enode://6f6c58bb00418b6892a59936717f597eb8a2e5113a5e68188bb958275f1c5f767fae2db4e9f675970c8b20d9f4da2a8e0066292725756446348eb6ea10ac6215@127.0.0.1:21002?discport=0",
	"enode://70465abd32f3bd107eb6cce1b99c57db154b98039efecfcb757b1b8244cbb26846025fe8a1eadfeaff9a7f49d624d86f82ade87b737233588f36fce78f3fa7b9@127.0.0.1:21003?discport=0"
]



genesis.json
{
    "config": {
        "chainId": 10,
        "homesteadBlock": 0,
        "eip150Block": 0,
        "eip150Hash": "0x0000000000000000000000000000000000000000000000000000000000000000",
        "eip155Block": 0,
        "eip158Block": 0,
        "byzantiumBlock": 0,
        "constantinopleBlock": 0,
        "petersburgBlock": 0,
        "istanbulBlock": 0,
        "istanbul": {
            "epoch": 30000,
            "policy": 0,
            "ceil2Nby3Block": 0,
            "testQBFTBlock": 0
        },
        "isQuorum": true,
        "txnSizeLimit": 64,
        "maxCodeSize": 0,
        "qip714Block": 0,
        "isMPS": false
    },
    "nonce": "0x0",
    "timestamp": "0x61251fb1",
    "extraData": "0xf87aa00000000000000000000000000000000000000000000000000000000000000000f8549482f114cde4898983626a27af4eb928ff804e60ee9421de2bf49c07595cf8c7c64ac5b173a112171cfe943f3e6c684f16f7fac8aa6ae20bb04a391036799494e63a320b26610685d3d5124c7c65b360735ab8f2c080c0",
    "gasLimit": "0xe0000000",
    "difficulty": "0x1",
    "mixHash": "0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365",
    "coinbase": "0x0000000000000000000000000000000000000000",
    "alloc": {
        "21de2bf49c07595cf8c7c64ac5b173a112171cfe": {
            "balance": "0x446c3b15f9926687d2c40534fdb564000000000000"
        },
        "3f3e6c684f16f7fac8aa6ae20bb04a3910367994": {
            "balance": "0x446c3b15f9926687d2c40534fdb564000000000000"
        },
        "82f114cde4898983626a27af4eb928ff804e60ee": {
            "balance": "0x446c3b15f9926687d2c40534fdb564000000000000"
        },
        "e63a320b26610685d3d5124c7c65b360735ab8f2": {
            "balance": "0x446c3b15f9926687d2c40534fdb564000000000000"
        }
    },
    "number": "0x0",
    "gasUsed": "0x0",
    "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000"
}
```

```
$ ./build/bin/qbft setup --help
NAME:
   qbft setup - Setup your qbft network in seconds

USAGE:
   qbft setup [command options] [arguments...]

DESCRIPTION:
   This tool helps generate:

    * Genesis block
    * Static nodes for all validators
    * Validator details

      for qbft consensus.


OPTIONS:
   --num value                Number of validators (default: 0)
   --nodes                    Print static nodes template
   --verbose                  Print validator details
   --quorum                   Use Quorum
   --save                     Save to files
   --nodeIp value             IP address of node (default: "0.0.0.0")
   --nodePortBase value       Base port number to use on node (default: 30303)
   --nodePortIncrement value  Value to increment port number by, for each node (default: 0)
```

</details>

### `address` subcommand

<details>
<summary>Click here to expand</summary>

This command is to extract Validator Address (ID) from node key hex which is the node private key in hex

E.g.:
```
$ ./build/bin/qbft address --nodekeyhex 1be3b50b31734be48452c29d714941ba165ef0cbf3ccea8ca16c45e3d8d45fb0
0xd8dba507e85f116b1f7e231ca8525fc9008a6966
```
</details>


## Testing

<details>
<summary>Click here to expand</summary>

### Integration tests

#### Istanbul BFT Geth Integration tests

* [Test specification](https://github.com/getamis/istanbul-tools/wiki/Istanbul-BFT-Test-Specification)
* [Source code](https://github.com/getamis/istanbul-tools/tree/develop/tests/functional)

#### Istanbul BFT Quorum Integration tests

* [Test specification](https://github.com/getamis/istanbul-tools/wiki/Istanbul-BFT-on-Quorum-Test-Specification)
* [Source code](https://github.com/getamis/istanbul-tools/tree/develop/tests/quorum/functional)

### Load tests

[Istanbul-BFT-Benchmarking](https://github.com/getamis/istanbul-tools/wiki/Istanbul-BFT-Benchmarking)

</details>
