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

package graphql

import (
	"fmt"
	"os"

	cobra "github.com/spf13/cobra"
	cmdbutils "github.com/universonic/ivy-utils/pkg/utils/cmdb"
)

// kickstartCmd represents the kickstart command
var kickstartCmd = &cobra.Command{
	Use:   "kickstart",
	Short: "Kickstart CMDB result generation",
	Long:  `Kickstart CMDB result generation. Ansible, and ansible-cmdb is required.`,
	Run: func(cmd *cobra.Command, args []string) {
		constructor := cmdbutils.NewResultConstructor()
		err := constructor.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return
		}
		err = constructor.ExportTo(outputDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return
		}
		fmt.Fprintf(os.Stdout, "Success\n")
	},
}

var outputDir string

func init() {
	cmdbCmd.AddCommand(kickstartCmd)
	cmdbCmd.PersistentFlags().StringVarP(
		&outputDir, "output-dir", "o", "./", "The directory to store combined result files",
	)
}
