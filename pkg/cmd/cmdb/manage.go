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
	storagecore "github.com/universonic/ivy-utils/pkg/storage/core"
	cmdbutil "github.com/universonic/ivy-utils/pkg/utils/cmdb"
)

// manageCmd represents the manage command
var manageCmd = &cobra.Command{
	Use:   "manage",
	Short: "Manage CMDB host entities",
	Long:  `Manage CMDB host entities.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		for _, each := range extraInfoOrig {
			kv := strings.Split(each, "=")
			if len(kv) != 2 {
				return fmt.Errorf("Invalid key-value pair: %s", each)
			}
			host.ExtraInfo[strings.Replace(kv[0], " ", "_", -1)] = kv[1]
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Validate parameters
		ok, isAction := validateActionFlags()
		if !ok {
			fmt.Fprintf(os.Stderr, "Multiple action flags was given.\n")
			os.Exit(1)
		}
		storage, err := NewStorageFromArgs()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not spawn storage due to: %v\n", err)
			os.Exit(10)
		}
		defer storage.Close()
		if hostComment != "" {
			host.ExtraInfo["comment"] = hostComment
		}
		if hostDept != "" {
			host.ExtraInfo["department"] = hostDept
		}
		inventory := cmdbutil.NewInventoryFromStorage(storage)
		if isAction {
			if len(args) == 0 {
				fmt.Fprintf(os.Stderr, "At least one hostname must be specified.\n")
				os.Exit(2)
			}
			host.Hostname = args[0]
			verify := func(hostname string) {
				host, err := inventory.Get(hostname)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not retrieve data from database with key '%s' due to: %v\n", hostname, err)
					os.Exit(12)
				}
				fmt.Fprintf(os.Stdout, "%s\n", host.CanonicalString())
			}
			if addHost {
				if host.IPMIAddress == "" || host.IPMIUser == "" || host.IPMIPassword == "" {
					fmt.Fprintf(os.Stderr, "IPMI endpoint and credential are required\n")
					os.Exit(2)
				}
				err = inventory.Add(*host)
			} else if removeHost {
				err = inventory.Delete(host.Hostname)
				if err == nil {
					fmt.Fprintf(os.Stdout, "Successfully deleted.\n")
					goto SKIP_VALIDATION
				}
			} else if updateHost {
				err = inventory.Update(*host)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not commit changes to database due to: %v\n", err)
				os.Exit(11)
			}
			verify(host.Hostname)
		SKIP_VALIDATION:
			return
		}
		var hosts []storagecore.Host
		if allHosts {
			hosts, err = inventory.List()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not retrieve data from database due to: %v\n", err)
				os.Exit(12)
			}
			goto FINALIZE
		} else if len(args) != 0 {
			for _, each := range args {
				host, err := inventory.Get(each)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not retrieve data from database with key '%s' due to: %v\n", args[0], err)
					os.Exit(12)
				}
				hosts = append(hosts, host)
			}
			goto FINALIZE
		} else {
			fmt.Fprintf(os.Stderr, "At least one host must be specified.\n")
			os.Exit(2)
		}
	FINALIZE:
		fmt.Fprintf(os.Stdout, "%s\n", storagecore.HostList(hosts).CanonicalString())
	},
}

var (
	host                                           = storagecore.NewHost()
	hostComment, hostDept                          string
	addHost, removeHost, updateHost, allHosts, yes bool
	extraInfoOrig                                  []string
)

func validateActionFlags() (ok, isAction bool) {
	actionFlags := 0
	if addHost {
		actionFlags++
	}
	if removeHost {
		actionFlags++
	}
	if updateHost {
		actionFlags++
	}
	return actionFlags <= 1, actionFlags > 0
}

func init() {
	cmdbCmd.AddCommand(manageCmd)

	manageCmd.Flags().BoolVarP(
		&addHost, "add", "a", addHost, "Add a new host",
	)
	manageCmd.Flags().BoolVarP(
		&removeHost, "remove", "r", removeHost, "Remove an existing host",
	)
	manageCmd.Flags().BoolVarP(
		&updateHost, "update", "u", updateHost, "Update an existing host",
	)
	manageCmd.Flags().BoolVar(
		&allHosts, "all", allHosts, "Select all existing hosts. It will be ignored if an action flag was specified.",
	)
	manageCmd.Flags().StringVar(
		&host.SSHAddress, "ssh-address", host.SSHAddress, "IP address that SSH service is listening on",
	)
	manageCmd.Flags().Uint16Var(
		&host.SSHPort, "ssh-port", 22, "Port of SSH service",
	)
	manageCmd.Flags().StringVar(
		&host.SSHUser, "ssh-user", "root", "Login user of SSH service",
	)
	manageCmd.Flags().StringVar(
		&host.IPMIAddress, "ipmi-address", "", "IP Address of IPMI interface",
	)
	manageCmd.Flags().StringVar(
		&host.IPMIUser, "ipmi-user", "", "Login user of IPMI interface",
	)
	manageCmd.Flags().StringVar(
		&host.IPMIPassword, "ipmi-password", "", "Login password of IPMI interface",
	)
	manageCmd.Flags().StringVar(
		&hostComment, "comment", "", "Comment of the node",
	)
	manageCmd.Flags().StringVar(
		&hostDept, "department", "", "Department of the node",
	)
	manageCmd.Flags().StringSliceVar(
		&extraInfoOrig, "extra-info", extraInfoOrig, "Comma-seperated key-value pair in 'key=value' format",
	)
}
