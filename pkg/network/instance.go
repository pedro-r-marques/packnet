/*
Copyright 2015 Juniper Networks, Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package network

import (
	"fmt"
	"strings"

	"github.com/Juniper/contrail-go-api"
	"github.com/Juniper/contrail-go-api/types"
)

type InstanceManager interface {
	LocateInstance(namespace, packName string) (*types.VirtualMachine, error)
	LocateInterface(network *types.VirtualNetwork, instance *types.VirtualMachine) (*types.VirtualMachineInterface, error)
	LocateInstanceIp(network *types.VirtualNetwork, nic *types.VirtualMachineInterface) (*types.InstanceIp, error)
	LocateInstanceGateway(network *types.VirtualNetwork) (string, error)
	LocateMacAddress(fqn string) (string, error)
}

type InstanceManagerImpl struct {
	client    contrail.ApiClient
	allocator AddressAllocator
}

func NewInstanceManager(client contrail.ApiClient, allocator AddressAllocator) InstanceManager {
	manager := new(InstanceManagerImpl)
	manager.client = client
	manager.allocator = allocator
	return manager
}

func instanceFQName(tenant, packName string) []string {
	fqn := []string{DefaultDomain, tenant, packName}
	return fqn
}

func (m *InstanceManagerImpl) LocateInstance(tenant, packName string) (*types.VirtualMachine, error) {
	fqn := instanceFQName(tenant, packName)
	instance, err := types.VirtualMachineByName(m.client, strings.Join(fqn, ":"))
	if err == nil && instance != nil {
		return instance, nil
	}

	instance = new(types.VirtualMachine)
	instance.SetFQName("project", fqn)
	err = m.client.Create(instance)
	if err != nil {
		log.Error("Create %s: %v", packName, err)
		return nil, err
	}
	return instance, nil
}

func (m *InstanceManagerImpl) DeleteInstance(uid string) error {
	err := m.client.DeleteByUuid("virtual-machine", uid)
	return err
}

func interfaceFQName(namespace, packName string) []string {
	fqn := []string{DefaultDomain, namespace, packName}
	return fqn
}

func (m *InstanceManagerImpl) LookupInterface(namespace, packName string) (*types.VirtualMachineInterface, error) {
	fqn := interfaceFQName(namespace, packName)
	ifc, err := types.VirtualMachineInterfaceByName(m.client, strings.Join(fqn, ":"))
	if err != nil {
		log.Error("Get vmi %s: %v", packName, err)
		return nil, err
	}
	return ifc, nil
}

func (m *InstanceManagerImpl) LocateInterface(network *types.VirtualNetwork, instance *types.VirtualMachine) (*types.VirtualMachineInterface, error) {
	namespace := instance.GetFQName()[len(instance.GetFQName())-2]
	fqn := interfaceFQName(namespace, instance.GetName())

	ifc, err := types.VirtualMachineInterfaceByName(m.client, strings.Join(fqn, ":"))
	if err == nil && ifc != nil {
		return ifc, nil
	}

	nic := new(types.VirtualMachineInterface)
	nic.SetFQName("project", fqn)
	nic.AddVirtualMachine(instance)
	if network != nil {
		nic.AddVirtualNetwork(network)
	}
	err = m.client.Create(nic)
	if err != nil {
		log.Error("Create interface %s: %v", instance.GetName(), err)
		return nil, err
	}

	_, err = types.VirtualMachineInterfaceByUuid(m.client, nic.GetUuid())
	if err != nil {
		log.Error("Get vmi %s: %v", nic.GetUuid(), err)
		return nil, err
	}
	return nic, nil
}

func (m *InstanceManagerImpl) ReleaseInterface(namespace, packName string) error {
	fqn := interfaceFQName(namespace, packName)
	vmi, err := types.VirtualMachineInterfaceByName(m.client, strings.Join(fqn, ":"))
	if err != nil {
		log.Error("Get vmi %s: %v", strings.Join(fqn, ":"), err)
		return err
	}

	refs, err := vmi.GetFloatingIpBackRefs()
	if err != nil {
		log.Error("Get %s floating-ip back refs: %v", packName, err)
		return err
	}
	for _, ref := range refs {
		err = m.client.DeleteByUuid("floating-ip", ref.Uuid)
		if err != nil {
			log.Error("Delete floating-ip %s: %v", ref.Uuid, err)
			return err
		}
	}

	err = m.client.Delete(vmi)
	if err != nil {
		log.Error("Delete vmi %s: %v", vmi.GetUuid(), err)
		return err
	}

	return nil
}

func makeInstanceIpName(tenant, nicName string) string {
	return tenant + "_" + nicName
}

func (m *InstanceManagerImpl) LocateInstanceIp(network *types.VirtualNetwork, nic *types.VirtualMachineInterface) (*types.InstanceIp, error) {
	tenant := nic.GetFQName()[len(nic.GetFQName())-2]
	ipName := makeInstanceIpName(tenant, nic.GetName())
	instanceIP, err := types.InstanceIpByName(m.client, ipName)
	if err == nil && instanceIP != nil {
		// TODO(prm): ensure that attributes are as expected
		return instanceIP, nil
	}

	address, err := m.allocator.LocateIpAddress(nic.GetUuid())
	if err != nil {
		return nil, err
	}

	// Create InstanceIp
	ipObj := &types.InstanceIp{}
	ipObj.SetName(ipName)
	ipObj.AddVirtualNetwork(network)
	ipObj.AddVirtualMachineInterface(nic)
	ipObj.SetInstanceIpAddress(address)
	err = m.client.Create(ipObj)
	if err != nil {
		log.Error("Create instance-ip %s: %v", nic.GetName(), err)
		return nil, err
	}

	_, err = m.client.FindByUuid(ipObj.GetType(), ipObj.GetUuid())
	if err != nil {
		log.Error("Get instance-ip %s: %v", ipObj.GetUuid(), err)
		return nil, err
	}
	return ipObj, nil
}

func (m *InstanceManagerImpl) ReleaseInstanceIp(namespace, nicName, instanceUID string) error {
	ipName := makeInstanceIpName(namespace, nicName)
	instanceIP, err := types.InstanceIpByUuid(m.client, ipName)
	if err != nil {
		log.Error("Get instance-ip %s: %v", ipName, err)
		return err
	}
	err = m.client.DeleteByUuid("instance-ip", instanceIP.GetSubnetUuid())
	if err != nil {
		log.Error("Delete instance-ip %s: %v", instanceIP.GetSubnetUuid(), err)
	}

	m.allocator.ReleaseIpAddress(instanceUID)
	return nil
}

func (m *InstanceManagerImpl) AttachFloatingIp(packName, projectName string, floatingIp *types.FloatingIp) error {
	fqn := append(strings.Split(projectName, ":"), packName)
	vmi, err := types.VirtualMachineInterfaceByName(m.client, strings.Join(fqn, ":"))
	if err != nil {
		log.Error("GET vmi %s: %v", packName, err)
		return err
	}

	refs, err := floatingIp.GetVirtualMachineInterfaceRefs()
	if err != nil {
		log.Error("GET floating-ip %s: %v", floatingIp.GetUuid(), err)
		return err
	}
	for _, ref := range refs {
		if ref.Uuid == vmi.GetUuid() {
			return nil
		}
	}

	floatingIp.AddVirtualMachineInterface(vmi)
	err = m.client.Update(floatingIp)
	if err != nil {
		log.Error("Update floating-ip %s: %v", packName, err)
		return err
	}
	return nil
}

func (m *InstanceManagerImpl) LocateInstanceGateway(network *types.VirtualNetwork) (string, error) {
	refs, err := network.GetNetworkIpamRefs()
	if err != nil {
		return "", fmt.Errorf("unable to retrieve network-ipam refs")
	}
	if len(refs) == 0 {
		return "", fmt.Errorf("no refs available.")
	}

	attr := refs[0].Attr.(types.VnSubnetsType)
	if len(attr.IpamSubnets) == 0 {
		return "", fmt.Errorf("IpamSubnets is empty.")
	}

	return attr.IpamSubnets[0].DefaultGateway, nil
}

func (m *InstanceManagerImpl) LocateMacAddress(fqn string) (string, error) {
	vmi, err := types.VirtualMachineInterfaceByName(m.client, fqn)
	if err != nil {
		log.Error("Get vmi %s: %v", fqn, err)
		return "", err
	}

	macs := vmi.GetVirtualMachineInterfaceMacAddresses()
	if len(macs.MacAddress) == 0 {
		return "", fmt.Errorf("no mac addresses found.")
	}

	return macs.MacAddress[0], nil
}
