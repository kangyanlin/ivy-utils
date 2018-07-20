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
	"sort"

	excel "github.com/360EntSecGroup-Skylar/excelize"
	core "github.com/universonic/ivy-utils/pkg/storage/core"
	cliutil "github.com/universonic/ivy-utils/pkg/utils/cli"
	zap "go.uber.org/zap"
)

const (
	DEF_XLSX_SHEET = "Overview"
)

type ReportMode byte

func (mod ReportMode) IsMode(mode ReportMode) bool {
	return mod == mode
}

func (mod ReportMode) Validate() error {
	var matched bool
	available := []ReportMode{ExportMode, AnsibleMode, JSONMode, XLSXMode, HTMLMode}
	for i := range available {
		rst := mod.IsMode(available[i])
		if rst {
			if matched {
				return ErrMultipleReportMode
			}
			matched = true
		}
	}
	if !matched {
		return ErrUnknownReportMode
	}
	return nil
}

const (
	ExportMode ReportMode = 0x01 << iota
	AnsibleMode
	JSONMode
	XLSXMode
	HTMLMode
)

var (
	ErrUnknownReportMode       = errors.New("Unknown report mode")
	ErrMultipleReportMode      = errors.New("Multiple report mode detected and only a single mode is acceptable")
	ErrUnmergableAnsibleResult = errors.New("Given ansible results' inventory did not match!")
)

type ReportGenerator struct {
	inventory *Inventory
}

func (in *ReportGenerator) GenerateAndSaveAs(selectedHosts []string, all bool, mode ReportMode, output string) (err error) {
	err = mode.Validate()
	if err != nil {
		return
	}
	sp := cliutil.NewSpinner()
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
		if err == nil {
			fmt.Fprintf(os.Stdout, "Accomplished!\n")
		} else {
			fmt.Fprintf(os.Stdout, "Something went wrong:\n")
		}
		sp.Stop()
	}()
	var (
		hosts      []core.Host
		numOfTasks int
	)
	if mode.IsMode(ExportMode) {
		numOfTasks = 1
	} else if mode.IsMode(HTMLMode) {
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
	if mode.IsMode(ExportMode) {
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
			cv0 := NewAnsibleResultMergableUnit()
			err := json.Unmarshal(input0[k], cv0)
			if err != nil {
				return nil, err
			}
			cv1 := NewAnsibleResultMergableUnit()
			err = json.Unmarshal(v, cv1)
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
	at0 := NewAnsibleTask("canonical", inventoryFile, true)
	at1 := NewAnsibleTask("ipmi", inventoryFile)
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
	// Detect if it is in JSONMode or XLSXMode
	if mode.IsMode(JSONMode) || mode.IsMode(XLSXMode) {
		var excel bool
		if mode.IsMode(XLSXMode) {
			excel = true
		}
		return in.QualifyReport(combinedData, excel, output)
	}
	// It is AnsibleMode or HTMLMode
	var outputDir string
	if mode.IsMode(HTMLMode) {
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
	if !mode.IsMode(HTMLMode) {
		return nil
	}
	printMsgOnStop(true)
	sp.Prefix = fmt.Sprintf("Generate html report (4/%d): ", numOfTasks)
	sp.Start()
	cmd := exec.Command("ansible-cmdb", "-i", inventoryFile, outputDir)
	tpl, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Could not generate html due to: %v\n\n%s", err, tpl)
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

func (in *ReportGenerator) QualifyReport(combinedData map[string][]byte, xlsx bool, output string) (err error) {
	var cvs []*AnsibleResultCarrier
	for _, each := range combinedData {
		cv := NewAnsibleResultCarrier()
		err = json.Unmarshal(each, &cv)
		if err != nil {
			return
		}
		cvs = append(cvs, cv)
	}
	var result []*QualifiedResult
	for i := range cvs {
		qr := NewQualifiedResult()
		qr.LoadFrom(cvs[i])
		result = append(result, qr)
	}
	sorter := NewQualifiedResultSorter(result)
	sort.Sort(sorter)
	result = sorter.Export()
	if xlsx {
		buf := excel.NewFile()
		buf.SetSheetName(buf.GetSheetName(1), DEF_XLSX_SHEET)
		/* ------------ WRITE TOP HEADER ------------ */
		buf.MergeCell(DEF_XLSX_SHEET, "A1", "E1")
		buf.SetCellStr(DEF_XLSX_SHEET, "A1", "Basic")
		buf.MergeCell(DEF_XLSX_SHEET, "F1", "H1")
		buf.SetCellStr(DEF_XLSX_SHEET, "F1", "Specification")
		buf.MergeCell(DEF_XLSX_SHEET, "I1", "J1")
		buf.SetCellStr(DEF_XLSX_SHEET, "I1", "Location")
		buf.MergeCell(DEF_XLSX_SHEET, "K1", "N1")
		buf.SetCellStr(DEF_XLSX_SHEET, "K1", "CPU")
		buf.MergeCell(DEF_XLSX_SHEET, "O1", "P1")
		buf.SetCellStr(DEF_XLSX_SHEET, "O1", "Memory")
		buf.MergeCell(DEF_XLSX_SHEET, "Q1", "R1")
		buf.SetCellStr(DEF_XLSX_SHEET, "Q1", "Disk")
		buf.MergeCell(DEF_XLSX_SHEET, "S1", "W1")
		buf.SetCellStr(DEF_XLSX_SHEET, "S1", "Network")
		/* ------------ WRITE PRIMARY HEADER ------------ */
		buf.MergeCell(DEF_XLSX_SHEET, "A2", "A3")
		buf.SetCellStr(DEF_XLSX_SHEET, "A2", "Name")
		buf.MergeCell(DEF_XLSX_SHEET, "B2", "B3")
		buf.SetCellStr(DEF_XLSX_SHEET, "B2", "OS")
		buf.MergeCell(DEF_XLSX_SHEET, "C2", "C3")
		buf.SetCellStr(DEF_XLSX_SHEET, "C2", "Department")
		buf.MergeCell(DEF_XLSX_SHEET, "D2", "D3")
		buf.SetCellStr(DEF_XLSX_SHEET, "D2", "Type")
		buf.MergeCell(DEF_XLSX_SHEET, "E2", "E3")
		buf.SetCellStr(DEF_XLSX_SHEET, "E2", "Comment")
		buf.MergeCell(DEF_XLSX_SHEET, "F2", "F3")
		buf.SetCellStr(DEF_XLSX_SHEET, "F2", "Manufacturer")
		buf.MergeCell(DEF_XLSX_SHEET, "G2", "G3")
		buf.SetCellStr(DEF_XLSX_SHEET, "G2", "Model")
		buf.MergeCell(DEF_XLSX_SHEET, "H2", "H3")
		buf.SetCellStr(DEF_XLSX_SHEET, "H2", "Serial Num")
		buf.MergeCell(DEF_XLSX_SHEET, "I2", "I3")
		buf.SetCellStr(DEF_XLSX_SHEET, "I2", "Rack Name")
		buf.MergeCell(DEF_XLSX_SHEET, "J2", "J3")
		buf.SetCellStr(DEF_XLSX_SHEET, "J2", "Rack Slot")
		buf.MergeCell(DEF_XLSX_SHEET, "K2", "K3")
		buf.SetCellStr(DEF_XLSX_SHEET, "K2", "Model")
		buf.MergeCell(DEF_XLSX_SHEET, "L2", "L3")
		buf.SetCellStr(DEF_XLSX_SHEET, "L2", "Base Freq")
		buf.MergeCell(DEF_XLSX_SHEET, "M2", "M3")
		buf.SetCellStr(DEF_XLSX_SHEET, "M2", "Count")
		buf.MergeCell(DEF_XLSX_SHEET, "N2", "N3")
		buf.SetCellStr(DEF_XLSX_SHEET, "N2", "Cores")
		buf.MergeCell(DEF_XLSX_SHEET, "O2", "O3")
		buf.SetCellStr(DEF_XLSX_SHEET, "O2", "DIMMs")
		buf.MergeCell(DEF_XLSX_SHEET, "P2", "P3")
		buf.SetCellStr(DEF_XLSX_SHEET, "P2", "Capacity")
		buf.MergeCell(DEF_XLSX_SHEET, "Q2", "R2")
		buf.SetCellStr(DEF_XLSX_SHEET, "Q2", "Virtual Disks")
		buf.SetCellStr(DEF_XLSX_SHEET, "Q3", "Name")
		buf.SetCellStr(DEF_XLSX_SHEET, "R3", "Size")
		buf.MergeCell(DEF_XLSX_SHEET, "S2", "S3")
		buf.SetCellStr(DEF_XLSX_SHEET, "S2", "Primary IP")
		buf.MergeCell(DEF_XLSX_SHEET, "T2", "T3")
		buf.SetCellStr(DEF_XLSX_SHEET, "T2", "IPMI Address")
		buf.MergeCell(DEF_XLSX_SHEET, "U2", "W2")
		buf.SetCellStr(DEF_XLSX_SHEET, "U2", "Logical Interfaces")
		buf.SetCellStr(DEF_XLSX_SHEET, "U3", "Name")
		buf.SetCellStr(DEF_XLSX_SHEET, "V3", "Member")
		buf.SetCellStr(DEF_XLSX_SHEET, "W3", "MAC")
		/* ------------ WRITE ROW ------------ */
		row := 4
		axisFactory := func(col rune, r int) string {
			return fmt.Sprintf("%s%d", string(col), r)
		}
		for _, qr := range result {
			lineHeight := 1
			buf.SetCellStr(DEF_XLSX_SHEET, axisFactory('A', row), qr.Name)
			buf.SetCellStr(DEF_XLSX_SHEET, axisFactory('B', row), qr.OS)
			buf.SetCellStr(DEF_XLSX_SHEET, axisFactory('C', row), qr.Department)
			buf.SetCellStr(DEF_XLSX_SHEET, axisFactory('D', row), qr.Type)
			buf.SetCellStr(DEF_XLSX_SHEET, axisFactory('E', row), qr.Comment)
			buf.SetCellStr(DEF_XLSX_SHEET, axisFactory('F', row), qr.Manufacturer)
			buf.SetCellStr(DEF_XLSX_SHEET, axisFactory('G', row), qr.Model)
			buf.SetCellStr(DEF_XLSX_SHEET, axisFactory('H', row), qr.SerialNumber)
			buf.SetCellStr(DEF_XLSX_SHEET, axisFactory('I', row), qr.RackName)
			buf.SetCellStr(DEF_XLSX_SHEET, axisFactory('J', row), qr.RackSlot)
			buf.SetCellStr(DEF_XLSX_SHEET, axisFactory('K', row), qr.CPUModel)
			buf.SetCellStr(DEF_XLSX_SHEET, axisFactory('L', row), qr.CPUBaseFreq)
			buf.SetCellValue(DEF_XLSX_SHEET, axisFactory('M', row), qr.CPUCount)
			buf.SetCellStr(DEF_XLSX_SHEET, axisFactory('N', row), fmt.Sprintf("%d / %d", qr.CPUCores, qr.CPUThreads))
			buf.SetCellStr(DEF_XLSX_SHEET, axisFactory('O', row), fmt.Sprintf("%d / %d", qr.PopulatedDIMMs, qr.MaximumDIMMs))
			buf.SetCellStr(DEF_XLSX_SHEET, axisFactory('P', row), qr.InstalledMemory)
			if h := len(qr.VirtualDisks); h > lineHeight {
				lineHeight = h
			}
			for vdi := range qr.VirtualDisks {
				buf.SetCellStr(DEF_XLSX_SHEET, axisFactory('Q', row+vdi), qr.VirtualDisks[vdi].Name)
				buf.SetCellStr(DEF_XLSX_SHEET, axisFactory('R', row+vdi), qr.VirtualDisks[vdi].Size)
			}
			buf.SetCellStr(DEF_XLSX_SHEET, axisFactory('S', row), qr.PrimaryIPAddress)
			buf.SetCellStr(DEF_XLSX_SHEET, axisFactory('T', row), qr.IPMIAddress)
			var h int
			for lii := range qr.LogicalInterfaces {
				length := len(qr.LogicalInterfaces[lii].Members)
				buf.MergeCell(DEF_XLSX_SHEET, axisFactory('U', row+h), axisFactory('U', row+h+length-1))
				buf.SetCellStr(DEF_XLSX_SHEET, axisFactory('U', row+h), qr.LogicalInterfaces[lii].Name)
				for limi := range qr.LogicalInterfaces[lii].Members {
					buf.SetCellStr(DEF_XLSX_SHEET, axisFactory('V', row+h+limi), qr.LogicalInterfaces[lii].Members[limi].Name)
					buf.SetCellStr(DEF_XLSX_SHEET, axisFactory('W', row+h+limi), qr.LogicalInterfaces[lii].Members[limi].MACAddress)
				}
				h += length
			}
			if h > lineHeight {
				lineHeight = h
			}
			if lineHeight > 1 {
				cols2align := "ABCDEFGHIJKLMNOPST"
				for _, each := range cols2align {
					buf.MergeCell(DEF_XLSX_SHEET, axisFactory(each, row), axisFactory(each, row+lineHeight-1))
				}
			}
			row += lineHeight
		}
		err = buf.SaveAs(output)
		return err
	}
	var dAtA []byte
	dAtA, err = json.Marshal(result)
	if err != nil {
		return
	}
	r := bytes.NewReader(dAtA)
	var fi *os.File
	fi, err = os.Create(output)
	if err != nil {
		return
	}
	_, err = io.Copy(fi, r)
	return err
}

func NewReportGenerator(storage core.Storage) *ReportGenerator {
	return &ReportGenerator{NewInventoryFromStorage(storage)}
}

type LogicalInterface struct {
	Name    string                    `json:"name"`
	Type    string                    `json:"type"`
	Members []*LogicalInterfaceMember `json:"members"`
}

func NewLogicalInterface() *LogicalInterface {
	return new(LogicalInterface)
}

type LogicalInterfaceMember struct {
	Name       string `json:"name"`
	MACAddress string `json:"mac_address"`
}

func NewLogicalInterfaceMember() *LogicalInterfaceMember {
	return new(LogicalInterfaceMember)
}

// QualifiedResult is a qualified struct of AnsibleResultCarrier. All
// fields are meant to be explicitly present in JSON output.
type QualifiedResult struct {
	Name              string              `json:"name"`
	OS                string              `json:"os"`
	Department        string              `json:"department"`
	Type              string              `json:"type"`
	Comment           string              `json:"comment"`
	Manufacturer      string              `json:"manufacturer"`
	Model             string              `json:"model"`
	SerialNumber      string              `json:"serial_number"`
	RackName          string              `json:"rack_name"`
	RackSlot          string              `json:"rack_slot"`
	CPUModel          string              `json:"cpu_model"`
	CPUBaseFreq       string              `json:"cpu_base_freq"`
	CPUCount          uint                `json:"cpu_count"`
	CPUCores          uint                `json:"cpu_cores"`
	CPUThreads        uint                `json:"cpu_threads"`
	PopulatedDIMMs    uint                `json:"populated_dimms"`
	MaximumDIMMs      uint                `json:"maximum_dimms"`
	InstalledMemory   string              `json:"installed_memory"`
	VirtualDisks      []*IPMIVirtualDisk  `json:"virtual_disks"`
	PhysicalDisks     []*IPMIPhysicalDisk `json:"physical_disks"`
	PrimaryIPAddress  string              `json:"primary_ip_address"`
	IPMIAddress       string              `json:"ipmi_address"`
	LogicalInterfaces []*LogicalInterface `json:"logical_intfs"`
}

func (in *QualifiedResult) LoadFrom(cv *AnsibleResultCarrier) {
	in.Name = cv.InventoryHostname
	if cv.Distribution == "OpenBSD" {
		in.OS = fmt.Sprintf("%s %s", cv.Distribution, cv.DistributionRelease)
	} else {
		in.OS = fmt.Sprintf("%s %s", cv.Distribution, cv.DistributionVersion)
	}
	in.Department = cv.Department
	switch cv.VirtualizationRole {
	case "NA", "host", "":
		in.Type = "Physical"
	case "?":
		in.Type = "(Not Sure)"
	default:
		in.Type = "Virtual"
	}
	in.Comment = cv.Comment
	in.Manufacturer = cv.IPMIManufacturer
	in.Model = cv.IPMIModel
	in.SerialNumber = cv.IPMISerialNumber
	if cv.IPMISystemLocation != nil {
		in.RackName = cv.IPMISystemLocation.RackName
		in.RackSlot = cv.IPMISystemLocation.RackSlot
	}
	cpuAllTheSame := true
	var (
		cpuModel             string
		cpuFreq              string
		cpuCores, cpuThreads uint
		cpuCount             uint
	)
	for _, each := range cv.IPMICPUs {
		if cpuModel == "" {
			cpuModel = each.Name
			cpuFreq = each.BaseClockSpeed
		}
		if each.Name != cpuModel {
			cpuAllTheSame = false
		}
		cpuCores += each.Cores
		cpuThreads += each.Threads
		cpuCount++
	}
	if !cpuAllTheSame {
		cpuModel += " and others"
	}
	in.CPUModel = cpuModel
	in.CPUBaseFreq = cpuFreq
	in.CPUCount = cpuCount
	in.CPUCores = cpuCores
	in.CPUThreads = cpuThreads
	in.PopulatedDIMMs = cv.IPMIPopulatedDIMMs
	in.MaximumDIMMs = cv.IPMIMaxDIMMs
	in.InstalledMemory = cv.IPMIMemoryInstalled
	in.VirtualDisks = cv.IPMIVirtualDisks
	in.PhysicalDisks = cv.IPMIPhysicalDisks
	if cv.DefaultIPv4 != nil {
		in.PrimaryIPAddress = cv.DefaultIPv4.Address
	}
	in.IPMIAddress = cv.IPMIAddress
	for k, v := range cv.Interfaces {
		intf := NewLogicalInterface()
		intf.Name = k
		intf.Type = v.Type
		switch intf.Type {
		case "bonding":
			for _, each := range v.Slaves {
				newMember := NewLogicalInterfaceMember()
				newMember.Name = each
				newMember.MACAddress = cv.Interfaces[each].MACAddress
				intf.Members = append(intf.Members, newMember)
			}
			in.LogicalInterfaces = append(in.LogicalInterfaces, intf)
		}
	}
}

func NewQualifiedResult() *QualifiedResult {
	return new(QualifiedResult)
}

type QualifiedResultSorter []*QualifiedResult

func (s QualifiedResultSorter) Len() int {
	return len(s)
}

func (s QualifiedResultSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s QualifiedResultSorter) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s QualifiedResultSorter) Export() []*QualifiedResult {
	return s
}

func NewQualifiedResultSorter(in []*QualifiedResult) QualifiedResultSorter {
	return in
}
