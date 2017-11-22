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

type Validator struct {
	Identity    int
	Genesis     string
	NodeKey     string
	StaticNodes string
	Port        int
	RPCPort     int
	IP          string
	EthStats    string
	Name        string
}

func NewValidator(identity int, genesis string, nodeKey string, staticNodes string, port int, rpcPort int, ethStats string, ip string) *Validator {
	return &Validator{
		Identity: identity,
		Genesis:  genesis,
		NodeKey:  nodeKey,
		Port:     port,
		RPCPort:  rpcPort,
		EthStats: ethStats,
		IP:       ip,
		Name:     fmt.Sprintf("validator-%v", identity),
	}
}

func (v Validator) String() string {
	tmpl, err := template.New("validator").Parse(validatorTemplate)
	if err != nil {
		fmt.Printf("Failed to parse template, %v", err)
		return ""
	}

	result := new(bytes.Buffer)
	err = tmpl.Execute(result, v)
	if err != nil {
		fmt.Printf("Failed to render template, %v", err)
		return ""
	}

	return result.String()
}

var validatorTemplate = `{{ .Name }}:
    hostname: {{ .Name }}
    image: quay.io/amis/geth:latest
    ports:
      - '{{ .Port }}:30303'
      - '{{ .RPCPort }}:8545'
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
    restart: always`
