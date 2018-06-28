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
	"path/filepath"

	cobra "github.com/spf13/cobra"
	cliutil "github.com/universonic/ivy-utils/pkg/utils/cli"
	cmdbutil "github.com/universonic/ivy-utils/pkg/utils/cmdb"
)

// kickstartCmd represents the kickstart command
var kickstartCmd = &cobra.Command{
	Use:        "kickstart",
	Short:      "Kickstart CMDB result generation",
	Long:       `Kickstart CMDB result generation. Ansible and iDrac plugin module is required.`,
	Deprecated: "use 'report' command instead. This command will be removed very soon.",
	Run: func(cmd *cobra.Command, args []string) {
		constructor := cmdbutil.NewResultConstructor()
		spinner := cliutil.NewSpinner()
		spinner.Suffix = "Collect host information: Processing... "
		spinner.Start()
		err := constructor.Run(inventoryFile)
		spinner.Stop()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Collect host information: Failed!\n\n\nReason:\n---\n%v\n---\n", err)
			os.Exit(2)
		} else {
			fmt.Fprintf(os.Stdout, "Collect host information: Completed!\n")
		}
		spinner = cliutil.NewSpinner()
		spinner.Suffix = "Export result: Processing... "
		spinner.Start()
		err = constructor.ExportTo(output)
		spinner.Stop()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Export result: Failed!\n\n\nReason:\n---\n%v\n---\n", err)
			os.Exit(3)
		} else {
			fmt.Fprintf(os.Stdout, "Export result: Completed!\n")
		}
		fmt.Fprintf(os.Stdout, "Success :)\n")
	},
}

func init() {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	currentDir := filepath.Dir(exe)

	cmdbCmd.AddCommand(kickstartCmd)

	kickstartCmd.PersistentFlags().StringVarP(
		&inventoryFile, "inventory-file", "i", "/etc/ansible/hosts", "The inventory file to be used by Ansible",
	)
	kickstartCmd.PersistentFlags().StringVarP(
		&output, "output-dir", "o", filepath.Join(currentDir, "output"), "The directory to store combined result files",
	)
}
