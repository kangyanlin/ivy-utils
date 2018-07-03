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
	"errors"
	"fmt"
	"os"

	tablewriter "github.com/olekukonko/tablewriter"
	core "github.com/universonic/ivy-utils/pkg/storage/core"
	cliutil "github.com/universonic/ivy-utils/pkg/utils/cli"
)

var ErrChangeNotCommited = errors.New("Modification succeeded but the change was not applied")

type HostLocationManager struct {
	hostID    string
	inventory *Inventory
}

func (in *HostLocationManager) Set(field, value string) (err error) {
	sp := cliutil.NewSpinner()
	var table *tablewriter.Table
	printMsgOnStop := func(succeeded bool) {
		if succeeded {
			fmt.Fprintf(os.Stdout, "\r"+sp.Prefix+"Completed!\n")
		} else {
			fmt.Fprintf(os.Stdout, "\r"+sp.Prefix+"Failed!\n")
		}
		sp.Stop()
	}
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
		printMsgOnStop(err == nil)
		if err != nil {
			fmt.Fprintf(os.Stdout, "Something went wrong:\n")
		} else if table != nil {
			table.Render()
		}
	}()
	sp.Prefix = "Modify node location (1/2): "
	sp.Start()
	var host core.Host
	host, err = in.inventory.Get(in.hostID)
	if err != nil {
		return
	}
	var params []string
	switch field {
	case "aisle":
		params = []string{"System", "Location", "Aisle"}
	case "datacenter":
		params = []string{"System", "Location", "DataCenter"}
	case "rackname":
		params = []string{"System", "Location", "Rack", "Name"}
	case "rackslot":
		params = []string{"System", "Location", "Rack", "Slot"}
	case "roomname":
		params = []string{"System", "Location", "RoomName"}
	default:
		return fmt.Errorf("No such field. Only 'aisle', 'datacenter', 'rackname', 'rackslot', and 'roomname' are acceptable")
	}
	setTask := NewRacadmCommandTask("set", host, value, params...)
	err = setTask.Execute()
	if err != nil {
		return
	}
	printMsgOnStop(true)
	sp.Prefix = "Verify changes (2/2): "
	sp.Start()
	getTask := NewRacadmCommandTask("get", host, "", "System", "Location")
	err = getTask.Execute()
	if err != nil {
		return
	}
	deviceSize, _ := getTask.Result["DeviceSize"]
	aisle, _ := getTask.Result["Aisle"]
	datacenter, _ := getTask.Result["DataCenter"]
	rackName, _ := getTask.Result["Rack.Name"]
	rackSlot, _ := getTask.Result["Rack.Slot"]
	roomName, _ := getTask.Result["RoomName"]
	switch field {
	case "aisle":
		if aisle != value {
			return ErrChangeNotCommited
		}
	case "datacenter":
		if datacenter != value {
			return ErrChangeNotCommited
		}
	case "rackname":
		if rackName != value {
			return ErrChangeNotCommited
		}
	case "rackslot":
		if rackSlot != value {
			return ErrChangeNotCommited
		}
	case "roomname":
		if roomName != value {
			return ErrChangeNotCommited
		}
	}
	table = tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"GUID", "Hostname", "Device Size", "Aisle", "Datacenter", "Rack Name", "Rack Slot", "Room Name"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	if deviceSize != "" {
		deviceSize += "U"
	}
	table.Append([]string{
		host.GUID,
		host.Hostname,
		deviceSize,
		aisle,
		datacenter,
		rackName,
		rackSlot,
		roomName,
	})
	return nil
}

func (in *HostLocationManager) Describe() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
		if err != nil {
			fmt.Fprintf(os.Stdout, "Something went wrong:\n")
		}
	}()
	var host core.Host
	host, err = in.inventory.Get(in.hostID)
	if err != nil {
		return
	}
	getTask := NewRacadmCommandTask("get", host, "", "System", "Location")
	err = getTask.Execute()
	if err != nil {
		return
	}
	deviceSize, _ := getTask.Result["DeviceSize"]
	aisle, _ := getTask.Result["Aisle"]
	datacenter, _ := getTask.Result["DataCenter"]
	rackName, _ := getTask.Result["Rack.Name"]
	rackSlot, _ := getTask.Result["Rack.Slot"]
	roomName, _ := getTask.Result["RoomName"]
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"GUID", "Hostname", "Device Size", "Aisle", "Datacenter", "Rack Name", "Rack Slot", "Room Name"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	if deviceSize != "" {
		deviceSize += "U"
	}
	table.Append([]string{
		host.GUID,
		host.Hostname,
		deviceSize,
		aisle,
		datacenter,
		rackName,
		rackSlot,
		roomName,
	})
	table.Render()
	return nil
}

func NewHostLocationManager(hostID string, storage core.Storage) *HostLocationManager {
	return &HostLocationManager{
		hostID:    hostID,
		inventory: NewInventoryFromStorage(storage),
	}
}
