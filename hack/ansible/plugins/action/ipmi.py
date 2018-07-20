from __future__ import (absolute_import, division, print_function)
__metaclass__ = type

from abc import ABCMeta, abstractmethod

import ssl
import urllib2
import json
import subprocess

from ansible.plugins.action import ActionBase

class ActionModule(ActionBase):
    '''IPMI Action Module for Ansible 2.3+
    '''
    _result = None
    _ipmi_addr = None
    _ipmi_user = None
    _ipmi_pass = None
    _manufacturer = None
    _model = None
    _serial_num = None

    def run(self, tmp=None, task_vars=None):

        if task_vars is None:
            task_vars = dict()
        self._result = AnsibleResult()
        result = super(ActionModule, self).run(tmp, task_vars)
        for k, v in result.items():
            self._result.__dict__[k] = v
        del tmp # tmp no longer has any effect

        # args parsers
        try:
            self._ipmi_addr = task_vars["ipmi_addr"]
        except KeyError:
            self._ipmi_addr = None
        try:
            self._ipmi_user = task_vars["ipmi_user"]
        except KeyError:
            self._ipmi_user = None
        try:
            self._ipmi_pass = task_vars["ipmi_pass"]
        except KeyError:
            self._ipmi_pass = None

        self._result.ansible_facts['ipmi_address'] = self._ipmi_addr
        adapter = self._initiate()
        if not adapter:
            return self._result.export()
        try:
            out = adapter.gather_info()
            for k, v in out.items():
                self._result.ansible_facts[k] = v
        except Exception as e:
            self._fail_with_message(e)

        self._result.ansible_facts['ansible_inventory_hostname'] = task_vars['inventory_hostname']
        try:
            self._result.ansible_facts['ansible_department'] = task_vars['department']
        except KeyError:
            self._result.ansible_facts['ansible_department'] = ''
        try:
            self._result.ansible_facts['ansible_comment'] = task_vars['comment']
        except KeyError:
            self._result.ansible_facts['ansible_comment'] = ''

        self._result.ansible_facts['ipmi_manufacturer'] = self._manufacturer
        self._result.ansible_facts['ipmi_model'] = self._model
        self._result.ansible_facts['ipmi_serial_num'] = self._serial_num
        
        return self._result.export()

    def _initiate(self):

        adapter = None
        if not self._ipmi_addr or not self._ipmi_user or not self._ipmi_pass:
            self._fail_with_message('Invalid IPMI endpoint or credential')
            return adapter

        def get_value(kv):
            return kv.split(':')[1].strip()

        lines = []
        failure = 0
        reason = None
        while failure < 3:
            try:
                lines = subprocess.check_output(['ipmitool', '-I', 'lanplus', '-H', self._ipmi_addr, '-U', self._ipmi_user, '-P', self._ipmi_pass, 'fru', 'print']).strip().split('\n')
                break
            except subprocess.CalledProcessError as e:
                if e.returncode == 1 and 'FRU' in e.output:
                    lines = e.output.strip().split('\n')
                    break
                reason = e
                failure += 1
        if failure >= 3:
            self._fail_with_message('Could not initialize IPMI adapter due to: %s' % reason)
            return adapter
        part_number = None
        product_name = None
        for line in lines:
            if line.startswith(' Product Manufacturer  :'):
                self._manufacturer = get_value(line)
                continue
            if line.startswith(' Product Part Number   :'):
                part_number = get_value(line)
                continue
            if line.startswith(' Product Name          :'):
                product_name = get_value(line)
                continue
            if line.startswith(' Product Serial        :'):
                self._serial_num = get_value(line)
                continue
            
        if self._manufacturer == 'DELL':
            adapter = DellAdapter(self._ipmi_addr, self._ipmi_user, self._ipmi_pass)
            self._model = product_name
        elif self._manufacturer == 'Supermicro':
            adapter = SupermicroAdapter(self._ipmi_addr, self._ipmi_user, self._ipmi_pass)
            self._model = part_number
        if not adapter:
            self._fail_with_message('Unsupported manufacturer: %s' % self._manufacturer)
        return adapter

    def _fail_with_message(self, msg=''):
        self._result.failed = True
        self._result.msg = str(msg) # => call '__str__' to stringify output

class AnsibleResult(object):
    ansible_facts = dict()
    changed = False
    failed = False
    msg = ''

    def export(self):
        result = dict()
        if self.failed:
            result['failed'] = self.failed
            result['msg'] = self.msg
        else:
            result['ansible_facts'] = self.ansible_facts
            result['changed'] = self.changed
        for k in [x for x in self.__dict__.keys() if not x in ['failed', 'msg', 'ansible_facts', 'changed']]:
            result[k] = self.__dict__[k]
        return result

class IPMIInterface:
    '''IPMIInterface is a generic interface for interaction with IPMI of various manufacturer.
    Multiple helpers are provided to make it easier to format output.
    '''
    __metaclass__ = ABCMeta

    def __init__(self, ipmi_addr, ipmi_user, ipmi_pass):
        self._ipmi_addr = ipmi_addr
        self._ipmi_user = ipmi_user
        self._ipmi_pass = ipmi_pass

    @abstractmethod
    def gather_info(self):
        return dict()

    @staticmethod
    def _fetch_from(key, dictionary):
        '''Peacefully get value of given key from given dict.
        '''
        v = None
        try:
            v = dictionary[key]
        except KeyError:
            v = None
        return v

    @staticmethod
    def _fmt_size(num, suffix='B'):
        '''Calculate given numeber of bytes to human-readable.
        '''
        for unit in ['','Ki','Mi','Gi','Ti','Pi','Ei','Zi']:
            if abs(num) < 1024.0:
                return "%3.1f %s%s" % (num, unit, suffix)
            num /= 1024.0
        return "%.1f %s%s" % (num, 'Yi', suffix)

class SupermicroAdapter(IPMIInterface):
    '''SupermicroAdapter is a adapter for Supermicro-based IPMI server
    '''

    def __init__(self, ipmi_addr, ipmi_user, ipmi_pass):
        super(SupermicroAdapter, self).__init__(ipmi_addr, ipmi_user, ipmi_pass)
        self._result = dict()

    def gather_info(self):
        # Unable to retrieve system location
        # FIXME: Currently we could not get any further details
        return self._result

class DellAdapter(IPMIInterface):
    '''DellAdapter is a adapter for Dell-based IPMI server
    '''

    def __init__(self, ipmi_addr, ipmi_user, ipmi_pass):
        super(DellAdapter, self).__init__(ipmi_addr, ipmi_user, ipmi_pass)
        self._result = dict()

    def gather_info(self):
        # TODO: use multi-threading process
        loc = self._get_system_location()
        location = dict()
        self._result['ipmi_system_location'] = location
        location['aisle'] = self._fetch_from("Aisle", loc)
        location['datacenter'] = self._fetch_from("DataCenter", loc)
        location['rack_name'] = self._fetch_from("Rack.Name", loc)
        location['rack_slot'] = self._fetch_from("Rack.Slot", loc)
        location['room_name'] = self._fetch_from("RoomName", loc)
        location['device_size'] = self._fetch_from("DeviceSize", loc)
        mem_settings = self._get_mem_settings()
        self._result['ipmi_installed_memory'] = self._fetch_from("SysMemSize", mem_settings)
        hwinventory = self._get_hardware_inventory()
        cpu_inventory = []
        self._result['ipmi_cpus'] = cpu_inventory
        if 'CPU' in hwinventory:
            for each in hwinventory['CPU'].values():
                if each['Device Type'] != 'CPU':
                    continue
                newItem = dict()
                try:
                    cpu = each['Model']
                    newItem['name'] = cpu.split('@')[0].strip()
                    newItem['manufacturer'] = each['Manufacturer']
                    newItem['cores'] = int(each['NumberOfEnabledCores'])
                    newItem['threads'] = int(each['NumberOfEnabledThreads'])
                    newItem['status'] = each['PrimaryStatus']
                    newItem['external_bus_clock_speed'] = each['ExternalBusClockSpeed']
                    newItem['base_clock_speed'] = each['CurrentClockSpeed']
                    newItem['turbo_clock_speed'] = each['MaxClockSpeed']
                except KeyError:
                    continue
                cpu_inventory.append(newItem)
        try:
            self._result['ipmi_maximum_dimms'] = int(hwinventory['System']['Embedded.1']['MaxDIMMSlots'])
            self._result['ipmi_populated_dimms'] = int(hwinventory['System']['Embedded.1']['PopulatedDIMMSlots'])
        except KeyError:
            self._result['ipmi_populated_dimms'] = None
            self._result['ipmi_maximum_dimms'] = None
        dimm_inventory = []
        self._result['ipmi_dimms'] = dimm_inventory
        if 'DIMM' in hwinventory:
            for each in hwinventory['DIMM'].values():
                newItem = dict()
                try:
                    newItem['name'] = each['DeviceDescription']
                    newItem['manufacturer'] = each['Manufacturer']
                    newItem['type'] = each['MemoryType']
                    newItem['model'] = each['Model']
                    newItem['part_number'] = each['PartNumber']
                    newItem['status'] = each['PrimaryStatus']
                    newItem['serial_number'] = each['SerialNumber']
                    newItem['size'] = each['Size']
                    newItem['speed'] = each['Speed']
                except KeyError:
                    continue
                dimm_inventory.append(newItem)
        vdisk_facts = []
        pdisk_facts = []
        self._result['ipmi_virtual_disks'] = vdisk_facts
        self._result['ipmi_physical_disks'] = pdisk_facts
        if 'Disk' in hwinventory:
            for each in hwinventory['Disk'].values():
                newDisk = dict()
                try:
                    if each['Device Type'] == 'PhysicalDisk':
                        newDisk['name'] = each['Model']
                        newDisk['description'] = each['DeviceDescription']
                        newDisk['status'] = each['PrimaryStatus']
                        newDisk['state'] = each['RaidStatus']
                        newDisk['serial_number'] = each['SerialNumber']
                        size = each['SizeInBytes'].split()
                        newDisk['size'] = self._fmt_size(int(size[0]))
                        newDisk['media_type'] = each['MediaType']
                        pdisk_facts.append(newDisk)
                    elif each['Device Type'] == 'VirtualDisk':
                        newDisk['name'] = each['Name']
                        newDisk['description'] = each['DeviceDescription']
                        newDisk['status'] = each['PrimaryStatus']
                        newDisk['state'] = each['RAIDStatus']
                        newDisk['layout'] = each['RAIDTypes']
                        size = each['SizeInBytes'].split()
                        newDisk['size'] = self._fmt_size(int(size[0]))
                        newDisk['media_type'] = each['MediaType']
                        vdisk_facts.append(newDisk)
                except KeyError:
                    continue
        return self._result

    def _get_system_location(self):
        return self.__get_racadm_kv(namespace="System.Location")
    
    def _get_mem_settings(self):
        return self.__get_racadm_kv(namespace="BIOS.MemSettings")

    def _get_hardware_inventory(self):
        output = self.__invoke_racadm('hwinventory', None)
        qualified_result = dict()
        started = False
        group = ''
        header = ''
        sections = -1
        for line in output:
            if not started:
                if line.startswith('----'):
                    started = True
                continue
            if line.startswith('[InstanceID: '):
                sections += 1
                title = line[13:len(line)-1]
                ss = title.split('.', 1)
                group = ss[0]
                header = ss[1]
                if not group in qualified_result:
                    qualified_result[group] = dict()
                if not header in qualified_result[group]:
                    qualified_result[group][header] = dict()
                continue
            if not line or not '=' in line:
                continue
            kv = line.split('=')
            k = kv[0].strip()
            v = kv[1].strip()
            qualified_result[group][header][k] = v
        return qualified_result
        
    def __get_racadm_kv(self, subcommand="get", namespace=None):
        output = self.__invoke_racadm(subcommand, namespace)
        qualified_output = dict()
        for i in output:
            if "=" in i:
                kv = i.split('=')
                k = kv[0]
                v = kv[1]
                if k.startswith('#'):
                    k = k[1:]
                qualified_output[k] = v
        return qualified_output

    def __invoke_racadm(self, subcommand, namespace=None, *args):
        cmd = ["racadm", "-r", self._ipmi_addr, "-u", self._ipmi_user, "-p", self._ipmi_pass, "--nocertwarn", subcommand]
        if namespace:
            cmd.append(namespace)
        for i in args:
            cmd.append(i)
        output = []
        try:
            output = subprocess.check_output(cmd).strip().split('\r\n')
        except subprocess.CalledProcessError as e:
            if e.returncode != 1:
                raise OSExecError(cmd, 'Hardware fault on IPMI interface')
            else:
                output = []
        if len(output) == 1:
            output = output[0].split('\n')
        return output

class OSExecError(Exception):
    '''OSExecError wraps given subprocess error inside as a new qualified error.
    '''
    def __init__(self, cmd, err):
        self._command = cmd
        self._error = err
    def __str__(self):
        return 'Failed to execute command "%s" due to: %s' % (' '.join(self._command), self._error)