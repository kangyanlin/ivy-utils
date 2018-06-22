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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

type QualifiedResult struct {
	RootDir string
	Data    map[string][]byte
}

func (in *QualifiedResult) Do(modulename string) error {
	cmd := exec.Command("ansible", "all", "-m", modulename, "-t", in.RootDir)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return in.sync()
}

func (in *QualifiedResult) sync() error {
	files, err := ioutil.ReadDir(in.RootDir)
	if err != nil {
		return err
	}
	defer os.RemoveAll(in.RootDir)
	for _, each := range files {
		if !each.IsDir() {
			fi, err := os.Open(filepath.Join(in.RootDir, each.Name()))
			if err != nil {
				return err
			}
			defer fi.Close()
			buf := new(bytes.Buffer)
			_, err = io.Copy(buf, fi)
			if err != nil {
				return err
			}
			in.Data[each.Name()] = buf.Bytes()
		}
	}
	return nil
}

func NewQualifiedResult() *QualifiedResult {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}
	return &QualifiedResult{
		RootDir: dir,
		Data:    make(map[string][]byte),
	}
}

type ResultCarrier struct {
	AnsibleFacts map[string]interface{} `json:"ansible_facts,omitempty"`
	Changed      bool                   `json:"changed,omitempty"`
}

type ResultConstructor struct {
	setupResult *QualifiedResult
	idracResult *QualifiedResult
	result      map[string][]byte
	wg          *sync.WaitGroup
}

func (in *ResultConstructor) Run() error {
	var err0, err1 error
	in.wg.Add(2)
	go func() {
		err0 = in.setupResult.Do("setup")
		in.wg.Done()
	}()
	go func() {
		err1 = in.idracResult.Do("idrac")
		in.wg.Done()
	}()
	in.wg.Wait()
	if err0 != nil {
		return fmt.Errorf("Could not collect basic dataset from setup module due to: %v", err0)
	}
	if err1 != nil {
		return fmt.Errorf("Could not collect basic dataset from idrac module due to: %v", err0)
	}
	dataset := make(map[string]ResultCarrier)
	for host, data := range in.setupResult.Data {
		var cv ResultCarrier
		err := json.Unmarshal(data, &cv)
		if err != nil {
			return err
		}
		var tmpCv ResultCarrier
		err = json.Unmarshal(in.idracResult.Data[host], &tmpCv)
		if err != nil {
			return err
		}
		cv.AnsibleFacts["idrac_address"] = tmpCv.AnsibleFacts["idrac_address"]
		cv.AnsibleFacts["idrac_model"] = tmpCv.AnsibleFacts["idrac_model"]
		cv.AnsibleFacts["idrac_bios_version"] = tmpCv.AnsibleFacts["idrac_bios_version"]
		cv.AnsibleFacts["idrac_bios_boot_mode"] = tmpCv.AnsibleFacts["idrac_bios_boot_mode"]
		cv.AnsibleFacts["idrac_hostname"] = tmpCv.AnsibleFacts["idrac_hostname"]
		cv.AnsibleFacts["idrac_system_location_aisle"] = tmpCv.AnsibleFacts["idrac_system_location_aisle"]
		cv.AnsibleFacts["idrac_system_location_datacenter"] = tmpCv.AnsibleFacts["idrac_system_location_datacenter"]
		cv.AnsibleFacts["idrac_system_location_rack_name"] = tmpCv.AnsibleFacts["idrac_system_location_rack_name"]
		cv.AnsibleFacts["idrac_system_location_rack_slot"] = tmpCv.AnsibleFacts["idrac_system_location_rack_slot"]
		cv.AnsibleFacts["idrac_system_location_room_name"] = tmpCv.AnsibleFacts["idrac_system_location_room_name"]
		cv.AnsibleFacts["idrac_device_size"] = tmpCv.AnsibleFacts["idrac_device_size"]
		cv.Changed = cv.Changed && tmpCv.Changed
		dataset[host] = cv
	}
	var err error
	for k, v := range dataset {
		in.result[k], err = json.Marshal(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (in *ResultConstructor) ExportTo(dirname string) error {
	err := os.MkdirAll(dirname, 0755)
	if err != nil {
		return err
	}
	for k, v := range in.result {
		buf := bytes.NewBuffer(v)
		fi, err := os.Create(filepath.Join(dirname, k))
		if err != nil {
			return err
		}
		defer fi.Close()
		_, err = io.Copy(fi, buf)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewResultConstructor() *ResultConstructor {
	wg := new(sync.WaitGroup)
	return &ResultConstructor{
		setupResult: NewQualifiedResult(),
		idracResult: NewQualifiedResult(),
		result:      make(map[string][]byte),
		wg:          wg,
	}
}
