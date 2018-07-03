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

	cobra "github.com/spf13/cobra"
	cmdbutil "github.com/universonic/ivy-utils/pkg/utils/cmdb"
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generates CMDB inventory report",
	Long:  `Generates CMDB inventory report. Ansible, Ansible iDrac plugin, Ansible canonical plugin is required.`,
	Run: func(cmd *cobra.Command, args []string) {
		storage, err := NewStorageFromArgs()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not spawn storage due to: %v\n", err)
			os.Exit(10)
		}
		defer storage.Close()
		generator := cmdbutil.NewReportGenerator(storage)
		err = generator.GenerateAndSaveAs(args, allHosts, inventoryOnly, html, output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not generate report due to: %v\n", err)
			os.Exit(20)
		}
	},
}

var (
	inventoryOnly, html bool
)

func init() {
	cmdbCmd.AddCommand(reportCmd)

	reportCmd.Flags().BoolVar(
		&inventoryOnly, "inventory-only", inventoryOnly, "Export inventory file only.",
	)
	reportCmd.Flags().StringVarP(
		&output, "output", "o", output, "Output file or directory.",
	)
	reportCmd.Flags().BoolVar(
		&html, "html", html, "Generate a static html report with ansible-cmdb.",
	)
	reportCmd.Flags().BoolVar(
		&allHosts, "all", allHosts, "Select all existing hosts.",
	)
}
