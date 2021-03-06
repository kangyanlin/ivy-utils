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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	core "github.com/universonic/ivy-utils/pkg/storage/core"
	zap "go.uber.org/zap"
)

type ParallelTasks struct {
	tasks    []Task
	errChan  chan error
	exportFn func() interface{}
	logger   *zap.SugaredLogger
}

func (in *ParallelTasks) Execute() (err error) {
	defer close(in.errChan)
	for _, task := range in.tasks {
		go func(t Task) {
			if e := t.Execute(); e != nil {
				in.errChan <- e
			} else {
				in.errChan <- nil
			}
		}(task)
	}
	var finished int
	for {
		if finished == len(in.tasks) {
			return err
		}
		select {
		case e := <-in.errChan:
			if e != nil && err == nil {
				err = e
			}
			finished++
		}
	}
}

func (in *ParallelTasks) GetResult() interface{} {
	if in.exportFn != nil {
		return in.exportFn()
	}
	return nil
}

func NewParallelTasks(tasks []Task, exportFn func() interface{}, logger *zap.SugaredLogger) *ParallelTasks {
	return &ParallelTasks{
		tasks:    tasks,
		errChan:  make(chan error, 8),
		exportFn: exportFn,
		logger:   logger,
	}
}

type Task interface {
	Execute() error
	GetResult() interface{}
}

type AnsibleTask struct {
	required      bool
	Module        string
	InventoryFile string
	Result        map[string][]byte
}

func (in *AnsibleTask) Execute() error {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}
	cmd := exec.Command("ansible", "all", "-m", in.Module, "-i", in.InventoryFile, "-t", dir)
	reason, err := cmd.Output()
	if err != nil && in.required {
		parsedReason := bytes.Split(reason, []byte("\n"))
		var finalReason []byte
		if len(parsedReason) > 16 {
			finalReason = bytes.Join(parsedReason[len(parsedReason)-17:], []byte("\n"))
		} else {
			finalReason = reason
		}
		return fmt.Errorf("error occurred while calling ansible module '%s': %v\n\n%s", in.Module, err, finalReason)
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	for _, each := range files {
		if !each.IsDir() {
			fi, err := os.Open(filepath.Join(dir, each.Name()))
			if err != nil {
				return err
			}
			defer fi.Close()
			buf := new(bytes.Buffer)
			_, err = io.Copy(buf, fi)
			if err != nil {
				return err
			}
			in.Result[each.Name()] = buf.Bytes()
		}
	}
	return nil
}

func (in *AnsibleTask) GetResult() interface{} {
	return in.Result
}

func NewAnsibleTask(moduleName, inventoryFile string, optional ...bool) *AnsibleTask {
	required := true
	if len(optional) != 0 && optional[0] {
		required = false
	}
	return &AnsibleTask{
		required:      required,
		Module:        moduleName,
		InventoryFile: inventoryFile,
		Result:        make(map[string][]byte),
	}
}

type InventoryExportTask struct {
	Hosts  []core.Host
	Result string
}

func (in *InventoryExportTask) Execute() (err error) {
	var rst string
	for _, each := range in.Hosts {
		var sshAddr string
		if each.SSHAddress == "" {
			sshAddr = each.Hostname
		} else {
			sshAddr = each.SSHAddress
		}
		rst += fmt.Sprintf("%s ansible_connection=\"smart\" ansible_host=\"%s\" ansible_port=%d ansible_user=\"%s\"",
			each.Hostname, sshAddr, each.SSHPort, each.SSHUser)
		if each.IPMIAddress != "" && each.IPMIUser != "" && each.IPMIPassword != "" {
			rst += fmt.Sprintf(" ipmi_addr=\"%s\" ipmi_user=\"%s\" ipmi_pass=\"%s\"", each.IPMIAddress, each.IPMIUser, each.IPMIPassword)
		}
		for k, v := range each.ExtraInfo {
			rst += fmt.Sprintf(" %s=\"%v\"", strings.Replace(k, " ", "_", -1), v)
		}
		rst += "\n"
	}
	fi, e := ioutil.TempFile("", "")
	if e != nil {
		return e
	}
	defer fi.Close()
	buf := bytes.NewBufferString(rst)
	_, err = io.Copy(fi, buf)
	if err != nil {
		return
	}
	in.Result = fi.Name()
	return nil
}

func (in *InventoryExportTask) GetResult() interface{} {
	return in.Result
}

func NewInventoryExportTask(hosts []core.Host) *InventoryExportTask {
	return &InventoryExportTask{
		Hosts: hosts,
	}
}

type RacadmCommandTask struct {
	Subcommand string
	Host       core.Host
	Namespace  []string
	Param      string
	Result     map[string]string
}

func (in *RacadmCommandTask) Execute() (err error) {
	if in.Host.IPMIAddress == "" {
		return fmt.Errorf("Given host's IPMI address was not allocated")
	}
	cmd := exec.Command("racadm", "-r", in.Host.IPMIAddress, "-u", in.Host.IPMIUser, "-p", in.Host.IPMIPassword, "--nocertwarn", in.Subcommand, strings.Join(in.Namespace, "."))
	if in.Param != "" {
		cmd.Args = append(cmd.Args, in.Param)
	}
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to execute command 'racadm' due to: %v", err)
	}
	out = bytes.TrimSpace(out)
	beautifiedOut := bytes.Split(out, []byte("\r\n"))
	for _, line := range beautifiedOut {
		if bytes.Contains(line, []byte("=")) {
			kv := bytes.Split(line, []byte("="))
			if len(kv) != 2 {
				continue
			}
			if bytes.HasPrefix(kv[0], []byte("#")) {
				kv[0] = bytes.TrimPrefix(kv[0], []byte("#"))
			}
			in.Result[string(kv[0])] = string(kv[1])
		}
	}
	return nil
}

func (in *RacadmCommandTask) GetResult() interface{} {
	return in.Result
}

func NewRacadmCommandTask(subcommand string, host core.Host, param string, namespace ...string) *RacadmCommandTask {
	return &RacadmCommandTask{
		Subcommand: subcommand,
		Host:       host,
		Param:      param,
		Namespace:  namespace,
		Result:     make(map[string]string),
	}
}
