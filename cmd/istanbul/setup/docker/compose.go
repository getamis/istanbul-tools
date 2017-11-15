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

package docker

import (
	"bytes"
	"fmt"
	"text/template"
)

type EthStats struct {
	Port   string
	Secret string
	IP     string
}

func (e EthStats) Stats() string {
	return fmt.Sprintf("%v@%v:%v", e.Secret, e.IP, e.Port)
}

type Compose struct {
	Services []Service
	IPPrefix string
	Stats    EthStats
}

func (c Compose) String() string {
	tmpl, err := template.New("compose").Parse(composeTemplate)
	if err != nil {
		fmt.Printf("Failed to parse template, %v", err)
		return ""
	}

	result := new(bytes.Buffer)
	err = tmpl.Execute(result, c)
	if err != nil {
		fmt.Printf("Failed to render template, %v", err)
		return ""
	}

	return result.String()
}

var composeTemplate = `version: '3'
services:
  eth-stats:
    image: quay.io/maicoin/eth-netstats:latest
    ports:
      - '3000:{{ .Stats.Port }}'
    environment:
      - WS_SECRET={{ .Stats.Secret }}
    restart: always
    networks:
      app_net:
        ipv4_address: {{ .Stats.IP }}
  {{- range .Services }}
  {{ . }}
  {{- end }}
networks:
  app_net:
    driver: bridge
    ipam:
      driver: default
      config:
      -
        subnet: {{ .IPPrefix }}.0/24`
