# istanbul-tools

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

## Getting started

```
$ make istanbul
$ build/bin/istanbul

NAME:
   istanbul - the istanbul-tools command line interface

USAGE:
   istanbul [global options] command [command options] [arguments...]

COMMANDS:
     decode   To decode an Istanbul extraData
     encode   To encode an Istanbul extraData
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help

COPYRIGHT:
   Copyright 2017 The Amis Authors

```

## Example

### Encode command

Encode the given file to `Encoded Istanbul extra-data` 
```
$ build/bin/istanbul encode --config ./cmd/istanbul/example/config.toml

OUTPUT:
Encoded Istanbul extra-data: 0x0000000000000000000000000000000000000000000000000000000000000000f89af85494475cc98b5521ab2a1335683e7567c8048bfe79ed9407d8299de61faed3686ba4c4e6c3b9083d7e2371944fe035ce99af680d89e2c4d73aca01dbfc1bd2fd94dc421209441a754f79c4a4ecd2b49c935aad0312b8410000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0
```

### Decode command

Decode extraData for the given input
```
$ build/bin/istanbul decode --extradata 0x0000000000000000000000000000000000000000000000000000000000000000f89af85494475cc98b5521ab2a1335683e7567c8048bfe79ed9407d8299de61faed3686ba4c4e6c3b9083d7e2371944fe035ce99af680d89e2c4d73aca01dbfc1bd2fd94dc421209441a754f79c4a4ecd2b49c935aad0312b8410000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0

OUTPUT:
vanity:  0x0000000000000000000000000000000000000000000000000000000000000000
validator:  0x475cc98b5521ab2a1335683e7567c8048bfe79ed
validator:  0x07d8299de61faed3686ba4c4e6c3b9083d7e2371
validator:  0x4fe035ce99af680d89e2c4d73aca01dbfc1bd2fd
validator:  0xdc421209441a754f79c4a4ecd2b49c935aad0312
seal: 0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
```