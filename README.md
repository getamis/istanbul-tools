# istanbul-tools

[![Test Status](https://travis-ci.org/getamis/istanbul-tools.svg?branch=feature%2Fadd-travis-yml)](https://travis-ci.org/getamis/istanbul-tools)


Istanbul tools contain tools for configuring `extraData` field for istanbul BFT (IBFT) network, integration tests for both IBFT Geth and Quorum, and load testing utilities for IBFT Geth.

Command line tools
---

### Build istanbul command line tools

```
$ make istanbul
$ build/bin/istanbul

NAME:
   istanbul - the istanbul-tools command line interface

USAGE:
   istanbul [global options] command [command options] [arguments...]

COMMANDS:
     extra    Istanbul extraData manipulation
     genesis  Istanbul genesis block generator
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help

COPYRIGHT:
   Copyright 2017 The Amis Authors

```

#### Extra data tool

Genesis extra-data encoder and decoder library for Istanbul consensus. 
   
istanbul-tools is used to generate extra-data field of genesis due to extra-data is combined signer vanity with RLP encoded `Istanbul extra data`. The `Istanbul extra data` struct is defined as follows:

```go
type IstanbulExtra struct {
    Validators    []common.Address  // Validator addresses
    Seal          []byte            // Proposer seal 65 bytes
    CommittedSeal [][]byte          // Committed seal, 65 * len(Validators) bytes
}
```

Please note: The `Seal`, and `CommittedSeal` is not considered in genesis block.  

##### Extra data subcommand
```
$ build/bin/istanbul extra

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

#####  Extra data example

###### Encode command

Encode the given file to `Encoded Istanbul extra-data` 
```
$ build/bin/istanbul extra encode --config ./cmd/istanbul/example/config.toml

OUTPUT:
Encoded Istanbul extra-data: 0x0000000000000000000000000000000000000000000000000000000000000000f89af85494475cc98b5521ab2a1335683e7567c8048bfe79ed9407d8299de61faed3686ba4c4e6c3b9083d7e2371944fe035ce99af680d89e2c4d73aca01dbfc1bd2fd94dc421209441a754f79c4a4ecd2b49c935aad0312b8410000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0
```

###### Decode command

Decode extraData for the given input
```
$ build/bin/istanbul extra decode --extradata 0x0000000000000000000000000000000000000000000000000000000000000000f89af85494475cc98b5521ab2a1335683e7567c8048bfe79ed9407d8299de61faed3686ba4c4e6c3b9083d7e2371944fe035ce99af680d89e2c4d73aca01dbfc1bd2fd94dc421209441a754f79c4a4ecd2b49c935aad0312b8410000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0

OUTPUT:
vanity:  0x0000000000000000000000000000000000000000000000000000000000000000
validator:  0x475cc98b5521ab2a1335683e7567c8048bfe79ed
validator:  0x07d8299de61faed3686ba4c4e6c3b9083d7e2371
validator:  0x4fe035ce99af680d89e2c4d73aca01dbfc1bd2fd
validator:  0xdc421209441a754f79c4a4ecd2b49c935aad0312
seal: 0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
```

##### Genesis subcommand
Genesis subcommand helps generating genesis.json. When `--nodes --verbose` flags are given, a `static-nodes.json` template as well as the validators' node keys, public keys, and addresses are generated.
**Note**: the generated `static-nodes.json` template are set with IP `0.0.0.0`, please make according change to match your environment.

###### Genesis tool example
```
 build/bin/istanbul genesis --num 4 --nodes --verbose                                                 ytlin@yt-macbook
{
	"Address": "0xf1b727db98587ee93ad3ac20631b0efb0fa5cef6",
	"Nodekey": "b623862cdb98efc608f5fb169718fe043fe5e5fe3e16ab24d4c4629dc32d28bf",
	"NodeInfo": "enode://49f11d1996c18e329857b84a0e1432145ff85f8709b57b3c79236eed824625b599f7f11904110319491c286c1ae3792095b39498e3465658914e5e3128c7dfe2@0.0.0.0:30303?discport=0"
}
{
	"Address": "0x1ebbedef9eebcaaac2aaaa078a79fea83f6b3256",
	"Nodekey": "075fc12664d3a1ad0ab6f79d4a48f66e4c67ed80253a63a6b69931d4801c95c9",
	"NodeInfo": "enode://071cad2c4cdbdf2946b47e65e0053c257c4f67a57b86c4cb26490c5f606be318df66e71fc4850fd22f097de3f8cca56ccb2e8669da02c4e2d6bb4abb017e76f8@0.0.0.0:30303?discport=0"
}
{
	"Address": "0xb490b19f328ec61e8d333e7781996561ee5c4210",
	"Nodekey": "4f5a57880196264bc1789bf050a3e7e2227f76ed3dd7d84542ba2eb8e418666b",
	"NodeInfo": "enode://457eef0b80276c48e5e630edd96c0e670733e58f088a5cb2150519ec1f50446b1d06e781c1bdc51e5b2f038427a2baf215f24d3340e31ec1bd22c10bdf00a6e0@0.0.0.0:30303?discport=0"
}
{
	"Address": "0x1be6bcd71ce0c73d44a2726cf364f4d57c3ee508",
	"Nodekey": "d05b431163372cf123ffe7fec6774cfe7f758ca5c7dcaadbb2d2a8d7f5a4c2df",
	"NodeInfo": "enode://c9f5ced64a0f2c96b99d8df148f9be07b56a7f4d773c3bca981c7247d6fc4e75c61a09ccb55f5d8c95fbad16c1515ce1d9b2b42c7f2991308ba2e090d3448633@0.0.0.0:30303?discport=0"
}

===========================================================

[
    "enode://49f11d1996c18e329857b84a0e1432145ff85f8709b57b3c79236eed824625b599f7f11904110319491c286c1ae3792095b39498e3465658914e5e3128c7dfe2@0.0.0.0:30303?discport=0",
    "enode://071cad2c4cdbdf2946b47e65e0053c257c4f67a57b86c4cb26490c5f606be318df66e71fc4850fd22f097de3f8cca56ccb2e8669da02c4e2d6bb4abb017e76f8@0.0.0.0:30303?discport=0",
    "enode://457eef0b80276c48e5e630edd96c0e670733e58f088a5cb2150519ec1f50446b1d06e781c1bdc51e5b2f038427a2baf215f24d3340e31ec1bd22c10bdf00a6e0@0.0.0.0:30303?discport=0",
    "enode://c9f5ced64a0f2c96b99d8df148f9be07b56a7f4d773c3bca981c7247d6fc4e75c61a09ccb55f5d8c95fbad16c1515ce1d9b2b42c7f2991308ba2e090d3448633@0.0.0.0:30303?discport=0"
]

===========================================================

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
    "timestamp": "0x59cdfa14",
    "extraData": "0x0000000000000000000000000000000000000000000000000000000000000000f89af85494f1b727db98587ee93ad3ac20631b0efb0fa5cef6941ebbedef9eebcaaac2aaaa078a79fea83f6b325694b490b19f328ec61e8d333e7781996561ee5c4210941be6bcd71ce0c73d44a2726cf364f4d57c3ee508b8410000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0",
    "gasLimit": "0x47b760",
    "difficulty": "0x1",
    "mixHash": "0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365",
    "coinbase": "0x0000000000000000000000000000000000000000",
    "alloc": {
        "1be6bcd71ce0c73d44a2726cf364f4d57c3ee508": {
            "balance": "0x446c3b15f9926687d2c40534fdb564000000000000"
        },
        "1ebbedef9eebcaaac2aaaa078a79fea83f6b3256": {
            "balance": "0x446c3b15f9926687d2c40534fdb564000000000000"
        },
        "b490b19f328ec61e8d333e7781996561ee5c4210": {
            "balance": "0x446c3b15f9926687d2c40534fdb564000000000000"
        },
        "f1b727db98587ee93ad3ac20631b0efb0fa5cef6": {
            "balance": "0x446c3b15f9926687d2c40534fdb564000000000000"
        }
    },
    "number": "0x0",
    "gasUsed": "0x0",
    "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000"
}
```

```
build/bin/istanbul genesis --help

NAME:
   istanbul genesis - Istanbul genesis block generator

USAGE:
   istanbul genesis [command options] [arguments...]

OPTIONS:
   --num value  Number of validators (default: 0)
   --nodes      Print static-nodes.json
   --verbose    Print more information
```

Integration tests
---
### Istanbul BFT Geth Integration tests
[Test specification](https://github.com/getamis/istanbul-tools/wiki/Istanbul-BFT-Test-Specification)
[Source code](https://github.com/getamis/istanbul-tools/tree/develop/tests/functional)

### Istanbul BFT Quorum Integration tests
[Test specification](https://github.com/getamis/istanbul-tools/wiki/Istanbul-BFT-on-Quorum-Test-Specification)
[Source code](https://github.com/getamis/istanbul-tools/tree/develop/tests/quorum/functional)


Load tests
---
[Test specification](https://github.com/getamis/istanbul-tools/wiki/Istanbul-BFT-Benchmarking)
