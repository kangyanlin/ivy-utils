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
	storage "github.com/universonic/ivy-utils/pkg/storage"
	storagecore "github.com/universonic/ivy-utils/pkg/storage/core"
	zap "go.uber.org/zap"
)

// cmdbCmd represents the cmdb command
var cmdbCmd = &cobra.Command{
	Use:   "cmdb",
	Short: "CMDB toolkit",
	Long:  `CMDB toolkit`,
}

var (
	inventoryFile string
	output        string
	// databaseType  string
	config        string
	configEnvName string
	configFile    string
)

// AttachTo attach subcommands onto parent command
func AttachTo(parentCmd *cobra.Command, parentParams ...interface{}) {
	parentCmd.AddCommand(cmdbCmd)
}

func NewStorageFromArgs() (storagecore.Storage, error) {
	logger := zap.NewNop().Sugar()
	if config != "" {
		return storage.NewStorageFromConfigBytes([]byte(config), logger)
	} else if configFile != "" {
		return storage.NewStorageFromConfigFile(configFile, logger)
	} else if configEnvName != "" {
		env := os.Getenv(configEnvName)
		return storage.NewStorageFromConfigBytes([]byte(env), logger)
	}
	return nil, fmt.Errorf("Database configuration not specified")
}

func init() {

	// cmdbCmd.PersistentFlags().StringVarP(
	// 	&databaseType, "database-adapter", "D", "etcd", "The adapter to use for database connection. Only 'etcd' is supported now.",
	// )
	cmdbCmd.PersistentFlags().StringVar(
		&config, "config", config, "Configuration of CMDB, which is in JSON format. This will override '--config-file'",
	)
	cmdbCmd.PersistentFlags().StringVarP(
		&configFile, "config-file", "c", configFile, "Configuration file of CMDB in YAML or JSON format. This will override '--config-env'",
	)
	cmdbCmd.PersistentFlags().StringVar(
		&configEnvName, "config-env", configEnvName, "Environment variable name to store configuration of CMDB. Its data must be in JSON format.",
	)
}
