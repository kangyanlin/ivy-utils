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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	core "github.com/universonic/ivy-utils/pkg/storage/core"
	cliutil "github.com/universonic/ivy-utils/pkg/utils/cli"
	zap "go.uber.org/zap"
)

var (
	ErrUnmergableAnsibleResult = errors.New("Given ansible results' inventory did not match!")
)

type AnsibleResultCarrier struct {
	AnsibleFacts map[string]interface{} `json:"ansible_facts,omitempty"`
	Changed      bool                   `json:"changed,omitempty"`
}

func NewAnsibleResultCarrier() *AnsibleResultCarrier {
	return new(AnsibleResultCarrier)
}

type ReportGenerator struct {
	inventory *Inventory
}

func (in *ReportGenerator) GenerateAndSaveAs(selectedHosts []string, all, inventoryOnly, html bool, output string) (err error) {
	sp := cliutil.NewSpinner()
	printMsgOnStop := func(succeeded bool) {
		if succeeded {
			fmt.Fprintf(os.Stderr, "\r"+sp.Prefix+"Completed!\n")
		} else {
			fmt.Fprintf(os.Stderr, "\r"+sp.Prefix+"Failed!\n")
		}
		sp.Stop()
	}
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
		printMsgOnStop(err == nil)
		if err == nil {
			fmt.Fprintf(os.Stderr, "Accomplished!\n")
		} else {
			fmt.Fprintf(os.Stderr, "Some thing went wrong:\n")
		}
		sp.Stop()
	}()
	var (
		hosts      []core.Host
		numOfTasks int
	)
	if inventoryOnly {
		numOfTasks = 1
	} else if html {
		numOfTasks = 4
	} else {
		numOfTasks = 3
	}
	sp.Prefix = fmt.Sprintf("Export inventory (1/%d): ", numOfTasks)
	sp.Start()
	if all {
		hosts, err = in.inventory.List()
		if err != nil {
			return err
		}
	} else {
		for _, host := range selectedHosts {
			out, err := in.inventory.Get(host)
			if err != nil {
				return err
			}
			hosts = append(hosts, out)
		}
	}
	inventoryTask := NewInventoryExportTask(hosts)
	err = inventoryTask.Execute()
	if err != nil {
		return err
	}
	inventoryFile := inventoryTask.GetResult().(string)
	defer os.Remove(inventoryFile)
	if inventoryOnly {
		to, err := os.Create(output)
		if err != nil {
			return err
		}
		defer to.Close()
		from, err := os.Open(inventoryFile)
		if err != nil {
			return err
		}
		defer from.Close()
		_, err = io.Copy(to, from)
		if err != nil {
			return err
		}
		return nil
	}
	merge := func(input0, input1 map[string][]byte) (map[string][]byte, error) {
		result := make(map[string][]byte)
		if len(input0) != len(input1) {
			return nil, ErrUnmergableAnsibleResult
		}
		for k := range input0 {
			if _, ok := input1[k]; !ok {
				return nil, ErrUnmergableAnsibleResult
			}
		}
		for k := range input1 {
			if _, ok := input0[k]; !ok {
				return nil, ErrUnmergableAnsibleResult
			}
		}
		for k, v := range input1 {
			var cv0 ResultCarrier
			err := json.Unmarshal(input0[k], &cv0)
			if err != nil {
				return nil, err
			}
			var cv1 ResultCarrier
			err = json.Unmarshal(v, &cv1)
			if err != nil {
				return nil, err
			}
			for k, v := range cv0.AnsibleFacts {
				cv1.AnsibleFacts[k] = v
			}
			cv1.Changed = cv0.Changed && cv1.Changed
			data, err := json.Marshal(cv1)
			if err != nil {
				return nil, err
			}
			result[k] = data
		}
		return result, nil
	}
	printMsgOnStop(true)
	sp.Prefix = fmt.Sprintf("Collect host information (2/%d): ", numOfTasks)
	sp.Start()
	at0 := NewAnsibleTask("canonical", inventoryFile)
	at1 := NewAnsibleTask("idrac", inventoryFile)
	ansibleTask := NewParallelTasks([]Task{at0, at1}, func() interface{} {
		result, err := merge(at0.Result, at1.Result)
		if err != nil {
			panic(err)
		}
		return result
	}, zap.NewNop().Sugar())
	err = ansibleTask.Execute()
	if err != nil {
		return err
	}
	combinedData := ansibleTask.GetResult().(map[string][]byte)
	printMsgOnStop(true)
	sp.Prefix = fmt.Sprintf("Export chunk data (3/%d): ", numOfTasks)
	sp.Start()
	var outputDir string
	if html {
		outputDir, err = ioutil.TempDir("", "")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(outputDir)
	} else {
		outputDir = output
	}
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return err
	}
	for k, v := range combinedData {
		buf := bytes.NewBuffer(v)
		fi, err := os.Create(filepath.Join(outputDir, k))
		if err != nil {
			return err
		}
		defer fi.Close()
		_, err = io.Copy(fi, buf)
		if err != nil {
			return err
		}
	}
	if !html {
		return nil
	}
	printMsgOnStop(true)
	sp.Prefix = fmt.Sprintf("Generate html report (4/%d): ", numOfTasks)
	sp.Start()
	cmd := exec.Command("ansible-cmdb", "-i", inventoryFile, outputDir)
	tpl, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Could not generate html due to: %v", err)
	}
	tmpReader := bytes.NewReader(tpl)
	outputHTML, err := os.Create(output)
	if err != nil {
		return err
	}
	defer outputHTML.Close()
	_, err = io.Copy(outputHTML, tmpReader)
	return err
}

func NewReportGenerator(storage core.Storage) *ReportGenerator {
	return &ReportGenerator{NewInventoryFromStorage(storage)}
}
