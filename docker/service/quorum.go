// Copyright 2017 AMIS Technologies
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package service

import (
	"bytes"
	"fmt"
	"text/template"
)

var (
	QuorumDockerImage    = "quorumengineering/quorum"
	QuorumDockerImageTag = "2.1.1"
)

type Quorum struct {
	*Validator
	Constellation *Constellation
}

func NewQuorum(v *Validator, c *Constellation) *Quorum {
	return &Quorum{
		Validator:     v,
		Constellation: c,
	}
}
func (q Quorum) String() string {
	tmpl, err := template.New("quorum").Parse(quorumTemplate)
	if err != nil {
		fmt.Printf("Failed to parse template, %v", err)
		return ""
	}

	result := new(bytes.Buffer)
	err = tmpl.Execute(result, q)
	if err != nil {
		fmt.Printf("Failed to render template, %v", err)
		return ""
	}

	return result.String()
}

var quorumTemplate = fmt.Sprintf(`{{ .Name }}:
    hostname: {{ .Name }}
    image: %s:%s
    ports:
      - '{{ .Port }}:30303'
      - '{{ .RPCPort }}:8545'
    volumes:
      - {{ .Identity }}:{{ .Constellation.Folder }}:z
    depends_on:
      - {{ .Constellation.Name }}
    environment:
      - PRIVATE_CONFIG={{ .Constellation.ConfigPath }}
    entrypoint:
      - /bin/sh
      - -c
      - |
        mkdir -p /eth/geth
        echo '{{ .Genesis }}' > /eth/genesis.json
        echo '{{ .StaticNodes }}' > /eth/geth/static-nodes.json
        geth --datadir "/eth" init "/eth/genesis.json"
        geth \
        --identity "{{ .Name }}" \
        --rpc \
        --rpcaddr "0.0.0.0" \
        --rpcport "8545" \
        --rpccorsdomain "*" \
        --datadir "/eth" \
        --port "30303" \
        --rpcapi "db,eth,net,web3,istanbul,personal" \
        --networkid "2017" \
        --nat "any" \
        --nodekeyhex "{{ .NodeKey }}" \
        --mine \
        --debug \
        --metrics \
        --syncmode "full" \
        --ethstats "{{ .Name }}:{{ .EthStats }}" \
        --gasprice 0
    networks:
      app_net:
        ipv4_address: {{ .IP }}
    restart: always
  {{ .Constellation }}`, QuorumDockerImage, QuorumDockerImageTag)
