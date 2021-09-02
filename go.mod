module github.com/Consensys/istanbul-tools

replace (
	github.com/ethereum/go-ethereum => github.com/Consensys/quorum v1.2.2-0.20210819085930-d5ef77cafd90
	github.com/ethereum/go-ethereum/crypto/secp256k1 => github.com/ConsenSys/quorum/crypto/secp256k1 v0.0.0-20210819085930-d5ef77cafd90
)

replace github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.1.1

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Microsoft/go-winio v0.4.11 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/PuerkitoBio/purell v1.0.0 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20160726150825-5bd2802263f2 // indirect
	github.com/aristanetworks/goarista v0.0.0-20181130030053-f7cbe917ef62 // indirect
	github.com/btcsuite/btcd v0.0.0-20181130015935-7d2daa5bfef2 // indirect
	github.com/cespare/cp v1.1.1 // indirect
	github.com/deckarep/golang-set v1.7.1 // indirect
	github.com/docker/distribution v0.0.0-20181129231500-d9e12182359e // indirect
	github.com/docker/docker v1.4.2-0.20180625184442-8e610b2b55bf
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.3.3 // indirect
	github.com/edsrzf/mmap-go v0.0.0-20170320065105-0bce6a688712 // indirect
	github.com/emicklei/go-restful v0.0.0-20170410110728-ff4f55a20633 // indirect
	github.com/emicklei/go-restful-swagger12 v0.0.0-20170208215640-dcef7f557305 // indirect
	github.com/ethereum/go-ethereum v1.8.27
	github.com/fjl/memsize v0.0.0-20190710130421-bcb5799ab5e5 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-openapi/analysis v0.0.0-20160815203709-b44dc874b601 // indirect
	github.com/go-openapi/jsonpointer v0.0.0-20160704185906-46af16f9f7b1 // indirect
	github.com/go-openapi/jsonreference v0.0.0-20160704190145-13c6e3589ad9 // indirect
	github.com/go-openapi/loads v0.0.0-20160704185440-18441dfa706d // indirect
	github.com/go-openapi/spec v0.0.0-20160808142527-6aced65f8501 // indirect
	github.com/go-openapi/swag v0.0.0-20160704191624-1d0bd113de87 // indirect
	github.com/google/uuid v1.1.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/howeyc/gopass v0.0.0-20170109162249-bf9dde6d0d2c // indirect
	github.com/imdario/mergo v0.3.6 // indirect
	github.com/inconshreveable/log15 v0.0.0-20180818164646-67afb5ed74ec
	github.com/influxdata/influxdb v1.7.7 // indirect
	github.com/juju/ratelimit v0.0.0-20170523012141-5b9ff8664717 // indirect
	github.com/mailru/easyjson v0.0.0-20160728113105-d5b7844b561a // indirect
	github.com/naoina/toml v0.1.2-0.20170918210437-9fafd6967416
	github.com/onsi/ginkgo v1.7.0
	github.com/onsi/gomega v1.4.3
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/prometheus/client_golang v1.5.1 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a
	github.com/rs/cors v0.0.0-20181011182759-a3460e445dd3 // indirect
	github.com/satori/go.uuid v1.1.0
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/ugorji/go v0.0.0-20170107133203-ded73eae5db7 // indirect
	github.com/urfave/cli v1.22.1
	go.uber.org/zap v1.14.1 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/olebedev/go-duktape.v3 v3.0.0-20190709231704-1e4459ed25ff // indirect
	k8s.io/apimachinery v0.0.0-20170728134514-1fd2e63a9a37
	k8s.io/client-go v4.0.0+incompatible
)

go 1.13
