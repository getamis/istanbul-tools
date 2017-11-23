# istanbul-tools

[![Test Status](https://travis-ci.org/getamis/istanbul-tools.svg?branch=feature%2Fadd-travis-yml)](https://travis-ci.org/getamis/istanbul-tools)

`istanbul-tools` contains tools for configuring Istanbul BFT (IBFT) network, integration tests for both IBFT Geth and Quorum, and load testing utilities for IBFT Geth.

## Build istanbul command line interface

```
$ make
$ ./build/bin/istanbul --help

NAME:
   istanbul - the istanbul-tools command line interface

USAGE:
   istanbul [global options] command [command options] [arguments...]

COMMANDS:
     extra    Istanbul extraData manipulation
     setup    Setup your Istanbul network in seconds
     help, h  Print this message

GLOBAL OPTIONS:
   --help, -h  show help

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

**Note**: the generated `static-nodes.json` template are set with IP `0.0.0.0`, please make according change to match your environment.

#### `setup` examples

```
$ ./build/bin/istanbul setup --num 4 --nodes --verbose
validators
{
    "Address": "0x5e5d0e2b80005a7e1f93044ddd64b2df0f8e488d",
    "Nodekey": "e5f9b868651ea8f4883744f2753ead9dfcdf7b1d8a96de0e733f406938dca1eb",
    "NodeInfo": "enode://8759a8a6921be78ec4e66ec77ae26ba9b3b1a51d1f83b16683c6f25e5a1d95a4de2c5bf4c2c05e1b984fae440236d96063efe933425df72659ee9de824cda6e1@0.0.0.0:30303?discport=0"
}
{
    "Address": "0x1b706dd850229813ee7c4002cd2fedc91380bb5a",
    "Nodekey": "2c13ee666b2ce617bf1e0d7fe7c8f058be27ea3a1aaabbfc63570a65f0bdae38",
    "NodeInfo": "enode://40dd1e7ba45e5bcd242420986d9d03133ce49399c6197e43254d523e94f547532d4c47c8aaba4b000c5a718568a48013b035c86f3ed8b13248888a15a76761c1@0.0.0.0:30303?discport=0"
}
{
    "Address": "0xdfdf27987b042bb3706d3a7c4b60e80a645744de",
    "Nodekey": "8bbf54eace8738f9d3ee90d5b949951f43d89acdb4b883d9188a141bdcd0153e",
    "NodeInfo": "enode://d188378b3eef56584b8ebd3da3ad579d39d23511943573cdeae5b8a37b5df22c369bf8900c4f42a9d4d5e55bc3cd357f319de8f833db3232295be22c8accc006@0.0.0.0:30303?discport=0"
}
{
    "Address": "0x5950b8f849daf1a78e119648c79111721353df59",
    "Nodekey": "9179c038483a2547c39f77f121065231d84a9c8d9bd044e1ddc19f653a23c751",
    "NodeInfo": "enode://d855be48593e6f2dd6201334e9381a2f01dac4a847385a393b1f664503b7b7020326e9f3f84f2d5713bf360d16566ed2b84d7df0b8b8313a7a4c4cf087ccfe27@0.0.0.0:30303?discport=0"
}



static-nodes.json
[
    "enode://8759a8a6921be78ec4e66ec77ae26ba9b3b1a51d1f83b16683c6f25e5a1d95a4de2c5bf4c2c05e1b984fae440236d96063efe933425df72659ee9de824cda6e1@0.0.0.0:30303?discport=0",
    "enode://40dd1e7ba45e5bcd242420986d9d03133ce49399c6197e43254d523e94f547532d4c47c8aaba4b000c5a718568a48013b035c86f3ed8b13248888a15a76761c1@0.0.0.0:30303?discport=0",
    "enode://d188378b3eef56584b8ebd3da3ad579d39d23511943573cdeae5b8a37b5df22c369bf8900c4f42a9d4d5e55bc3cd357f319de8f833db3232295be22c8accc006@0.0.0.0:30303?discport=0",
    "enode://d855be48593e6f2dd6201334e9381a2f01dac4a847385a393b1f664503b7b7020326e9f3f84f2d5713bf360d16566ed2b84d7df0b8b8313a7a4c4cf087ccfe27@0.0.0.0:30303?discport=0"
]



genesis.json
{
    "config": {
        "chainId": 2017,
        "homesteadBlock": 1,
        "eip150Block": 2,
        "eip150Hash": "0x0000000000000000000000000000000000000000000000000000000000000000",
        "eip155Block": 3,
        "eip158Block": 3,
        "istanbul": {
            "epoch": 30000,
            "policy": 0
        }
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

[Test specification](https://github.com/getamis/istanbul-tools/wiki/Istanbul-BFT-Benchmarking)

</details>