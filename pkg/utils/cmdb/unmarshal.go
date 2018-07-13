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

import "encoding/json"

type AnsibleAppArmor struct {
	Status string `json:"status,omitempty"`
}

func NewAnsibleAppArmor() *AnsibleAppArmor {
	return new(AnsibleAppArmor)
}

type AnsibleDateTime struct {
	Date              string `json:"date,omitempty"`
	Day               string `json:"day,omitempty"`
	Epoch             string `json:"epoch,omitempty"`
	Hour              string `json:"hour,omitempty"`
	ISO8601           string `json:"iso8601,omitempty"`
	ISO8601Basic      string `json:"iso8601_basic,omitempty"`
	ISO8601BasicShort string `json:"iso8601_basic_short,omitempty"`
	ISO8601Micro      string `json:"iso8601_micro,omitempty"`
	Minute            string `json:"minute,omitempty"`
	Month             string `json:"month,omitempty"`
	Second            string `json:"second,omitempty"`
	Time              string `json:"time,omitempty"`
	TimeZone          string `json:"tz,omitempty"`
	TimeZoneOffset    string `json:"tz_offset,omitempty"`
	Weekday           string `json:"weekday,omitempty"`
	WeekdayNumber     string `json:"weekday_number,omitempty"`
	WeekNumber        string `json:"weeknumber,omitempty"`
	Year              string `json:"year,omitempty"`
}

func NewAnsibleDateTime() *AnsibleDateTime {
	return new(AnsibleDateTime)
}

type AnsibleDefaultIP struct {
	Address    string `json:"address,omitempty"`
	Alias      string `json:"alias,omitempty"`
	Broadcast  string `json:"broadcast,omitempty"`
	Gateway    string `json:"gateway,omitempty"`
	Interface  string `json:"interface,omitempty"`
	MACAddress string `json:"macaddress,omitempty"`
	MTU        uint   `json:"mtu,omitempty"`
	Netmask    string `json:"netmask,omitempty"`
	Network    string `json:"network,omitempty"`
	Type       string `json:"type,omitempty"`
}

func NewAnsibleDefaultIP() *AnsibleDefaultIP {
	return new(AnsibleDefaultIP)
}

type AnsibleDeviceLinks struct {
	IDs     map[string][]string `json:"ids,omitempty"`
	Labels  map[string][]string `json:"labels,omitempty"`
	Masters map[string][]string `json:"masters,omitempty"`
	UUIDs   map[string][]string `json:"uuids,omitempty"`
}

func NewAnsibleDeviceLinks() *AnsibleDeviceLinks {
	return &AnsibleDeviceLinks{
		IDs:     make(map[string][]string),
		Labels:  make(map[string][]string),
		Masters: make(map[string][]string),
		UUIDs:   make(map[string][]string),
	}
}

type AnsiblePartition struct {
	Holders    []string `json:"holders,omitempty"`
	Host       string   `json:"host,omitempty"`
	Sectors    string   `json:"sectors,omitempty"`
	SectorSize uint16   `json:"sectorsize,omitempty"`
	Size       string   `json:"size,omitempty"`
	Start      string   `json:"start,omitempty"`
	UUID       string   `json:"uuid,omitempty"`
}

func NewAnsiblePartition() *AnsiblePartition {
	return new(AnsiblePartition)
}

type AnsibleDevice struct {
	Holders         []string                     `json:"holders,omitempty"`
	Host            string                       `json:"host,omitempty"`
	Links           map[string][]string          `json:"links,omitempty"`
	Model           string                       `json:"model,omitempty"`
	Partitions      map[string]*AnsiblePartition `json:"partitions,omitempty"`
	Removable       string                       `json:"removable,omitempty"`
	Rotational      string                       `json:"rotational,omitempty"`
	SASAddress      string                       `json:"sas_address,omitempty"`
	SASDeviceHandle string                       `json:"sas_device_handle,omitempty"`
	SchedulerMode   string                       `json:"scheduler_mode,omitempty"`
	Sectors         string                       `json:"sectors,omitempty"`
	Sectorsize      string                       `json:"sectorsize,omitempty"`
	Size            string                       `json:"size,omitempty"`
	SupportDiscard  string                       `json:"support_discard,omitempty"`
	Vendor          string                       `json:"vendor,omitempty"`
	Virtual         uint16                       `json:"virtual,omitempty"`
	WWN             string                       `json:"wwn,omitempty"`
}

func NewAnsibleDevice() *AnsibleDevice {
	return &AnsibleDevice{
		Links:      make(map[string][]string),
		Partitions: make(map[string]*AnsiblePartition),
	}
}

type AnsibleDNS struct {
	Domain      string   `json:"domain,omitempty"`
	NameServers []string `json:"nameservers,omitempty"`
	Search      []string `json:"search,omitempty"`
}

func NewAnsibleDNS() *AnsibleDNS {
	return new(AnsibleDNS)
}

type AnsibleIPv4Address struct {
	Address   string `json:"address,omitempty"`
	Broadcast string `json:"broadcast,omitempty"`
	Netmask   string `json:"netmask,omitempty"`
	Network   string `json:"network,omitempty"`
}

func NewAnsibleIPv4Address() *AnsibleIPv4Address {
	return new(AnsibleIPv4Address)
}

type AnsibleIPv6Address struct {
	Address string `json:"address,omitempty"`
	Prefix  string `json:"prefix,omitempty"`
	Scope   string `json:"scope,omitempty"`
}

func NewAnsibleIPv6Address() *AnsibleIPv6Address {
	return new(AnsibleIPv6Address)
}

type AnsibleNetworkInterface struct {
	// TODO: lack of some fields of bridge
	Active             bool                  `json:"active,omitempty"`
	Device             string                `json:"device,omitempty"`
	Features           map[string]string     `json:"features,omitempty"`
	HWTimestampFilters []string              `json:"hw_timestamp_filters,omitempty"`
	LACPRate           string                `json:"lacp_rate,omitempty"` // bonding
	IPv4               *AnsibleIPv4Address   `json:"ipv4,omitempty"`
	IPv6               []*AnsibleIPv6Address `json:"ipv6,omitempty"`
	MACAddress         string                `json:"macaddress,omitempty"`
	Miimon             string                `json:"miimon,omitempty"` // bonding
	Mode               string                `json:"mode,omitempty"`   // bonding
	Module             string                `json:"module,omitempty"` // ether
	MTU                uint32                `json:"mtu,omitempty"`
	PCIID              string                `json:"pciid,omitempty"`     // ether
	PHCIndex           int                   `json:"phc_index,omitempty"` // ether
	Promisc            bool                  `json:"promisc,omitempty"`
	Slaves             []string              `json:"slaves,omitempty"`     // bonding
	Interfaces         []string              `json:"interfaces,omitempty"` // bridge
	Speed              int                   `json:"speed,omitempty"`      // bonding
	Timestamping       []string              `json:"timestamping,omitempty"`
	Type               string                `json:"type,omitempty"`
}

type AnsibleLV struct {
	SizeG string `json:"size_g,omitempty"`
	VG    string `json:"vg,omitempty"`
}

type AnsiblePV struct {
	FreeG string `json:"free_g,omitempty"`
	SizeG string `json:"size_g,omitempty"`
	VG    string `json:"vg,omitempty"`
}

type AnsibleVG struct {
	FreeG string `json:"free_g,omitempty"`
	LVs   string `json:"num_lvs,omitempty"`
	PVs   string `json:"num_pvs,omitempty"`
	SizeG string `json:"size_g,omitempty"`
}

type AnsibleLVM struct {
	LVs map[string]*AnsibleLV `json:"lvs,omitempty"`
	PVs map[string]*AnsiblePV `json:"pvs,omitempty"`
	VGs map[string]*AnsibleVG `json:"vgs,omitempty"`
}

type AnsibleNoCacheMemory struct {
	Free uint `json:"free,omitempty"`
	Used uint `json:"used,omitempty"`
}
type AnsibleRealMemory struct {
	Free  uint `json:"free,omitempty"`
	Total uint `json:"total,omitempty"`
	Used  uint `json:"used,omitempty"`
}
type AnsibleSwapMemory struct {
	Cached uint `json:"cached,omitempty"`
	Free   uint `json:"free,omitempty"`
	Total  uint `json:"total,omitempty"`
	Used   uint `json:"used,omitempty"`
}

type AnsibleMemoryMB struct {
	NoCache *AnsibleNoCacheMemory `json:"nocache,omitempty"`
	Real    *AnsibleRealMemory    `json:"real,omitempty"`
	Swap    *AnsibleSwapMemory    `json:"swap,omitempty"`
}

type AnsibleMount struct {
	BlockAvailable uint   `json:"block_available,omitempty"`
	BlockSize      uint   `json:"block_size,omitempty"`
	BlockTotal     uint   `json:"block_total,omitempty"`
	BlockUsed      uint   `json:"block_used,omitempty"`
	Device         string `json:"device,omitempty"`
	FSType         string `json:"fstype,omitempty"`
	InodeAvailable uint   `json:"inode_available,omitempty"`
	InodeTotal     uint   `json:"inode_total,omitempty"`
	InodeUsed      uint   `json:"inode_used,omitempty"`
	Mount          string `json:"mount,omitempty"`
	Options        string `json:"options,omitempty"`
	SizeAvailable  uint   `json:"size_available,omitempty"`
	SizeTotal      uint   `json:"size_total,omitempty"`
	UUID           string `json:"uuid,omitempty"`
}

type AnsibleSELinux struct {
	Status string `json:"status,omitempty"`
}

type IPMIDIMM struct {
	Manufacturer string `json:"manufacturer,omitempty"`
	Model        string `json:"model,omitempty"`
	Name         string `json:"name,omitempty"`
	PartNumber   string `json:"part_number,omitempty"`
	SerialNumber string `json:"serial_number,omitempty"`
	Size         string `json:"size,omitempty"`
	Speed        string `json:"speed,omitempty"`
	Status       string `json:"status,omitempty"`
	Type         string `json:"type,omitempty"`
}

type IPMIVirtualDisk struct {
	Description string `json:"description,omitempty"`
	Layout      string `json:"layout,omitempty"`
	MediaType   string `json:"media_type,omitempty"`
	Name        string `json:"name,omitempty"`
	Size        string `json:"size,omitempty"`
	State       string `json:"state,omitempty"`
	Status      string `json:"status,omitempty"`
}

type IPMIPhysicalDisk struct {
	Description  string `json:"description,omitempty"`
	MediaType    string `json:"media_type,omitempty"`
	Name         string `json:"name,omitempty"`
	SerialNumber string `json:"serial_number,omitempty"`
	Size         string `json:"size,omitempty"`
	State        string `json:"state,omitempty"`
	Status       string `json:"status,omitempty"`
}

type AnsibleFacts struct {
	AllIPv4Addresses           []string                            `json:"ansible_all_ipv4_addresses,omitempty"`
	AllIPv6Addresses           []string                            `json:"ansible_all_ipv6_addresses,omitempty"`
	AppArmor                   *AnsibleAppArmor                    `json:"ansible_apparmor,omitempty"`
	Architecture               string                              `json:"ansible_architecture,omitempty"`
	BIOSDate                   string                              `json:"ansible_bios_date,omitempty"`
	BIOSVersion                string                              `json:"ansible_bios_version,omitempty"`
	Cmdline                    map[string]interface{}              `json:"ansible_cmdline,omitempty"`
	DateTime                   *AnsibleDateTime                    `json:"ansible_date_time,omitempty"`
	DefaultIPv4                *AnsibleDefaultIP                   `json:"ansible_default_ipv4,omitempty"`
	DefaultIPv6                *AnsibleDefaultIP                   `json:"ansible_default_ipv6,omitempty"`
	DeviceLinks                *AnsibleDeviceLinks                 `json:"ansible_device_links,omitempty"`
	Devices                    map[string]*AnsibleDevice           `json:"ansible_devices,omitempty"`
	Distribution               string                              `json:"ansible_distribution,omitempty"`
	DistributionFileParsed     bool                                `json:"ansible_distribution_file_parsed,omitempty"`
	DistributionFilePath       string                              `json:"ansible_distribution_file_path,omitempty"`
	DistributionFileVariety    string                              `json:"ansible_distribution_file_variety,omitempty"`
	DistributionMajorVersion   string                              `json:"ansible_distribution_major_version,omitempty"`
	DistributionRelease        string                              `json:"ansible_distribution_release,omitempty"`
	DistributionVersion        string                              `json:"ansible_distribution_version,omitempty"`
	DNS                        *AnsibleDNS                         `json:"ansible_dns,omitempty"`
	Domain                     string                              `json:"ansible_domain,omitempty"`
	EffectiveGroupID           uint16                              `json:"ansible_effective_group_id,omitempty"`
	EffectiveUserID            uint16                              `json:"ansible_effective_user_id,omitempty"`
	EnvironmentVariables       map[string]string                   `json:"ansible_env,omitempty"`
	FIPs                       bool                                `json:"ansible_fips,omitempty"`
	FormFactor                 string                              `json:"ansible_form_factor,omitempty"`
	FQDN                       string                              `json:"ansible_fqdn,omitempty"`
	Hostname                   string                              `json:"ansible_hostname,omitempty"`
	Interfaces                 map[string]*AnsibleNetworkInterface `json:"ansible_interfaces,omitempty"`
	IsChroot                   bool                                `json:"ansible_is_chroot,omitempty"`
	Kernel                     string                              `json:"ansible_kernel,omitempty"`
	AnsibleLocal               map[string]interface{}              `json:"ansible_local,omitempty"`
	LSB                        map[string]string                   `json:"ansible_lsb,omitempty"`
	LVM                        *AnsibleLVM                         `json:"ansible_lvm,omitempty"`
	Machine                    string                              `json:"ansible_machine,omitempty"`
	MachineID                  string                              `json:"ansible_machine_id,omitempty"`
	MemFreeMB                  uint                                `json:"ansible_memfree_mb,omitempty"`
	MemTotalMB                 uint                                `json:"ansible_memtotal_mb,omitempty"`
	MemoryMB                   *AnsibleMemoryMB                    `json:"ansible_memory_mb,omitempty"`
	Mounts                     []*AnsibleMount                     `json:"ansible_mounts,omitempty"`
	NodeName                   string                              `json:"ansible_nodename,omitempty"`
	OSFamily                   string                              `json:"ansible_os_family,omitempty"`
	PackageManager             string                              `json:"ansible_pkg_mgr,omitempty"`
	Processor                  []string                            `json:"ansible_processor,omitempty"`
	ProcessorCores             uint                                `json:"ansible_processor_cores,omitempty"`
	ProcessorCount             uint                                `json:"ansible_processor_count,omitempty"`
	ProcessorThreadsPerCore    uint                                `json:"ansible_processor_threads_per_core,omitempty"`
	ProcessorVCPUs             uint                                `json:"ansible_processor_vcpus,omitempty"`
	ProductName                string                              `json:"ansible_product_name,omitempty"`
	ProductSerial              string                              `json:"ansible_product_serial,omitempty"`
	ProductUUID                string                              `json:"ansible_product_uuid,omitempty"`
	ProductVersion             string                              `json:"ansible_product_version,omitempty"`
	Python                     map[string]interface{}              `json:"ansible_python,omitempty"`
	PythonVersion              string                              `json:"ansible_python_version,omitempty"`
	RealGroupID                uint16                              `json:"ansible_real_group_id,omitempty"`
	RealUserID                 uint16                              `json:"ansible_real_user_id,omitempty"`
	SELinux                    *AnsibleSELinux                     `json:"ansible_selinux,omitempty"`
	SELinuxPythonPresent       bool                                `json:"ansible_selinux_python_present,omitempty"`
	ServiceManager             string                              `json:"ansible_service_mgr,omitempty"`
	SSHHostKeyDSAPublic        string                              `json:"ansible_ssh_host_key_dsa_public,omitempty"`
	SSHHostKeyRSAPublic        string                              `json:"ansible_ssh_host_key_rsa_public,omitempty"`
	SwapFreeMB                 uint                                `json:"ansible_swapfree_mb,omitempty"`
	SwapTotalMB                uint                                `json:"ansible_swaptotal_mb,omitempty"`
	System                     string                              `json:"ansible_system,omitempty"`
	SystemCapabilities         []interface{}                       `json:"ansible_system_capabilities,omitempty"`
	SystemCapabilitiesEnforced string                              `json:"ansible_system_capabilities_enforced,omitempty"`
	SystemVendor               string                              `json:"ansible_system_vendor,omitempty"`
	UptimeSeconds              uint                                `json:"ansible_uptime_seconds,omitempty"`
	UserDir                    string                              `json:"ansible_user_dir,omitempty"`
	UserGecos                  string                              `json:"ansible_user_gecos,omitempty"`
	UserGID                    uint16                              `json:"ansible_user_gid,omitempty"`
	UserID                     string                              `json:"ansible_user_id,omitempty"`
	UserShell                  string                              `json:"ansible_user_shell,omitempty"`
	UserUID                    uint16                              `json:"ansible_user_uid,omitempty"`
	UserspaceArchitecture      string                              `json:"ansible_userspace_architecture,omitempty"`
	UserspaceBits              string                              `json:"ansible_userspace_bits,omitempty"`
	VirtualizationRole         string                              `json:"ansible_virtualization_role,omitempty"`
	VirtualizationType         string                              `json:"ansible_virtualization_type,omitempty"`
	GatherSubset               []string                            `json:"gather_subset,omitempty"`
	// IPMI module entities
	IPMIType            string              `json:"ipmi_type,omitempty"`
	IPMIAddress         string              `json:"ipmi_address,omitempty"`
	IPMIModel           string              `json:"ipmi_model,omitempty"`
	IPMIBiosVersion     string              `json:"ipmi_bios_version,omitempty"`
	IPMIBiosBootMode    string              `json:"ipmi_bios_boot_mode,omitempty"`
	IPMIHostname        string              `json:"ipmi_hostname,omitempty"`
	IPMIAisle           string              `json:"ipmi_system_location_aisle,omitempty"`
	IPMIDatacenter      string              `json:"ipmi_system_location_datacenter,omitempty"`
	IPMIRackName        string              `json:"ipmi_system_location_rack_name,omitempty"`
	IPMIRackSlot        string              `json:"ipmi_system_location_rack_slot,omitempty"`
	IPMIRoomName        string              `json:"ipmi_system_location_room_name,omitempty"`
	IPMIDeviceSize      string              `json:"ipmi_device_size,omitempty"`
	IPMIDIMMs           []*IPMIDIMM         `json:"ipmi_dimms,omitempty"`
	IPMIMemoryInstalled string              `json:"ipmi_installed_memory,omitempty"`
	IPMIMaxDIMMs        uint                `json:"ipmi_maximum_dimms,omitempty"`
	IPMIPopulatedDIMMs  uint                `json:"ipmi_populated_dimms,omitempty"`
	IPMIVirtualDisks    []*IPMIVirtualDisk  `json:"ipmi_virtual_disks,omitempty"`
	IPMIPhysicalDisks   []*IPMIPhysicalDisk `json:"ipmi_physical_disks,omitempty"`
	// Those are duplicated from hostvars
	Department string `json:"ansible_department,omitempty"`
	Comment    string `json:"ansible_comment,omitempty"`
}

func NewAnsibleFacts() *AnsibleFacts {
	return &AnsibleFacts{
		Cmdline:              make(map[string]interface{}),
		Devices:              make(map[string]*AnsibleDevice),
		EnvironmentVariables: make(map[string]string),
		Interfaces:           make(map[string]*AnsibleNetworkInterface),
		AnsibleLocal:         make(map[string]interface{}),
		LSB:                  make(map[string]string),
		Python:               make(map[string]interface{}),
	}
}

type AnsibleResultCarrier struct {
	*AnsibleFacts `json:"ansible_facts,omitempty"`
	Changed       bool `json:"changed,omitempty"`
}

func NewAnsibleResultCarrier() *AnsibleResultCarrier {
	return &AnsibleResultCarrier{
		AnsibleFacts: NewAnsibleFacts(),
	}
}

type AnsibleResultMergableUnit struct {
	AnsibleFacts map[string]interface{} `json:"ansible_facts,omitempty"`
	Changed      bool                   `json:"changed,omitempty"`
}

func (in *AnsibleResultMergableUnit) LoadFrom(data []byte) error {
	err := json.Unmarshal(data, in)
	if err != nil {
		return err
	}
	return nil
}

func NewAnsibleResultMergableUnit() *AnsibleResultMergableUnit {
	return &AnsibleResultMergableUnit{
		AnsibleFacts: make(map[string]interface{}),
	}
}
