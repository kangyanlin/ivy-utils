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

package cmdb

import (
	"fmt"
	"net"

	core "github.com/universonic/ivy-utils/pkg/storage/core"
)

type Inventory struct {
	Storage core.Storage
}

func (in *Inventory) Add(host core.Host) error {
	if host.SSHAddress != "" {
		ip := net.ParseIP(host.SSHAddress)
		if ip == nil {
			return fmt.Errorf("Invalid IP address: %s", host.SSHAddress)
		}
	}
	_, ok := host.ExtraInfo["comment"]
	if !ok {
		host.ExtraInfo["comment"] = ""
	}
	return in.Storage.CreateHost(host)
}

func (in *Inventory) Get(hostID string) (core.Host, error) {
	return in.Storage.GetHost(hostID)
}

func (in *Inventory) List() ([]core.Host, error) {
	return in.Storage.ListHost()
}

func (in *Inventory) Update(host core.Host) error {
	if host.SSHAddress != "" {
		ip := net.ParseIP(host.SSHAddress)
		if ip == nil {
			return fmt.Errorf("Invalid IP address: %s", host.SSHAddress)
		}
	}
	return in.Storage.UpdateHost(host.Hostname, func(h core.Host) (core.Host, error) {
		host.GUID = h.GUID
		if host.SSHAddress == "" {
			host.SSHAddress = h.SSHAddress
		}
		if host.SSHPort == 0 {
			host.SSHPort = h.SSHPort
		}
		if host.SSHUser == "" {
			host.SSHUser = h.SSHUser
		}
		if host.IPMIAddress == "" {
			host.IPMIAddress = h.IPMIAddress
		}
		if host.IPMIUser == "" {
			host.IPMIUser = h.IPMIUser
		}
		if host.IPMIPassword == "" {
			host.IPMIPassword = h.IPMIPassword
		}
		if _, ok := h.ExtraInfo["comment"]; !ok {
			h.ExtraInfo["comment"] = ""
		}
		if len(host.ExtraInfo) == 0 {
			host.ExtraInfo = h.ExtraInfo
		} else if _, ok := host.ExtraInfo["comment"]; !ok {
			host.ExtraInfo["comment"] = h.ExtraInfo["comment"]
		}
		return host, nil
	})
}

func (in *Inventory) Delete(hostID string) error {
	return in.Storage.DeleteHost(hostID)
}

func NewInventoryFromStorage(storage core.Storage) *Inventory {
	return &Inventory{storage}
}
