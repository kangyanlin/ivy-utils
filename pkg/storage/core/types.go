// Copyright © 2018 Alfred Chou <unioverlord@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"bytes"
	"encoding/json"
	"strconv"

	tablewriter "github.com/olekukonko/tablewriter"
)

// Host indicates host data object
type Host struct {
	GUID         string           `json:"guid,omitempty" yaml:"guid,omitempty"`
	Hostname     string           `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	SSHAddress   string           `json:"ssh_addr,omitempty" yaml:"ssh_addr,omitempty"`
	SSHPort      uint16           `json:"ssh_port,omitempty" yaml:"ssh_addr,omitempty"`
	SSHUser      string           `json:"ssh_user,omitempty" yaml:"ssh_user,omitempty"`
	IPMIAddress  string           `json:"ipmi_addr,omitempty" yaml:"ipmi_addr,omitempty"`
	IPMIUser     string           `json:"ipmi_user,omitempty" yaml:"ipmi_user,omitempty"`
	IPMIPassword string           `json:"ipmi_pass,omitempty" yaml:"ipmi_pass,omitempty"`
	ExtraInfo    ExtendableFields `json:"extra_info,omitempty" yaml:"extra_info,omitempty"`
}

func (host Host) CanonicalString() string {
	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"GUID", "Hostname", "SSH Address", "SSH Port", "SSH User", "IPMI Address", "IPMI User", "IPMI Password", "Extra Info"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	str, err := json.Marshal(host.ExtraInfo)
	if err != nil {
		str = []byte("<N/A>")
	}
	var armoredPassword string
	if host.IPMIPassword != "" {
		armoredPassword = "******"
	}
	table.Append([]string{
		host.GUID,
		host.Hostname,
		host.SSHAddress,
		strconv.Itoa(int(host.SSHPort)),
		host.SSHUser,
		host.IPMIAddress,
		host.IPMIUser,
		armoredPassword,
		string(str),
	})
	table.Render()
	return buf.String()
}

func NewHost() *Host {
	return &Host{
		ExtraInfo: make(ExtendableFields),
	}
}

// ExtendableFields is a general purpose extendable field mapping
type ExtendableFields map[string]interface{}

type HostList []Host

func (hosts HostList) CanonicalString() string {
	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"GUID", "Hostname", "SSH Address", "SSH Port", "SSH User", "IPMI Address", "IPMI User", "IPMI Password", "Extra Info"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	for _, host := range hosts {
		str, err := json.Marshal(host.ExtraInfo)
		if err != nil {
			str = []byte("<N/A>")
		}
		table.Append([]string{
			host.GUID,
			host.Hostname,
			host.SSHAddress,
			strconv.Itoa(int(host.SSHPort)),
			host.SSHUser,
			host.IPMIAddress,
			host.IPMIUser,
			"******",
			string(str),
		})
	}
	table.Render()
	return buf.String()
}
