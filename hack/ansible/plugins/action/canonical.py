from __future__ import (absolute_import, division, print_function)
__metaclass__ = type

from ansible.plugins.action import ActionBase

class ActionModule(ActionBase):
    '''Canonical Action Module for Ansible
    '''

    def run(self, tmp=None, task_vars=None):

        if task_vars is None:
            task_vars = dict()

        result = super(ActionModule, self).run(tmp, task_vars)

        new_module_args = self._task.args.copy()

        result.update(self._execute_module(module_name="setup", module_args=new_module_args, task_vars=task_vars, tmp=tmp, delete_remote_tmp=False))

        network_interfaces = dict()
        if result['ansible_facts'].has_key("ansible_interfaces"):
            for intf in result['ansible_facts']["ansible_interfaces"]:
                network_interfaces[intf] = result['ansible_facts'].pop("ansible_%s" % intf.replace("-", "_"))
        result['ansible_facts']['ansible_interfaces'] = network_interfaces
        
        self._remove_tmp_path(tmp)
        return result
