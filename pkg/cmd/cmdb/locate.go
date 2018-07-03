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
	"os"
	"strings"

	cobra "github.com/spf13/cobra"
	cmdbutil "github.com/universonic/ivy-utils/pkg/utils/cmdb"
)

// locateCmd represents the locate command
var locateCmd = &cobra.Command{
	Use:   "locate",
	Short: "Locate a specified host from inventory",
	Long:  `Locate a specified host from inventory`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Fprintf(os.Stderr, "Only a single host must be specified in arguments\n")
			os.Exit(2)
		}
		hostID := args[0]
		storage, err := NewStorageFromArgs()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not spawn storage due to: %v\n", err)
			os.Exit(10)
		}
		defer storage.Close()
		if modify {
			if location == "" {
				fmt.Fprintf(os.Stderr, "'--modify' flag must be used with '--location KEY=VALUE'\n")
				os.Exit(1)
			}
			kv := strings.Split(location, "=")
			if len(kv) != 2 {
				fmt.Fprintf(os.Stderr, "Invalid node location key-value pair\n")
				os.Exit(1)
			}
			manager := cmdbutil.NewHostLocationManager(hostID, storage)
			err = manager.Set(kv[0], kv[1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not modify host location due to: %v\n", err)
				os.Exit(20)
			}
			return
		}
		manager := cmdbutil.NewHostLocationManager(hostID, storage)
		err = manager.Describe()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not locate host due to: %v\n", err)
			os.Exit(20)
		}
	},
}

var (
	location string
	modify   bool
)

func init() {
	cmdbCmd.AddCommand(locateCmd)

	locateCmd.Flags().StringVarP(
		&location, "location", "l", location, "Key-value pair in 'key=value' format. Acceptable: aisle, datacenter, rackname, rackslot, and roomname",
	)
	locateCmd.Flags().BoolVarP(
		&modify, "modify", "m", modify, "Modify a host's location, must be used with '--location'",
	)
}
