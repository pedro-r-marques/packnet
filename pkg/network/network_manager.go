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

	"github.com/golang/glog"

	"github.com/Juniper/contrail-go-api"
	"github.com/Juniper/contrail-go-api/config"
	"github.com/Juniper/contrail-go-api/types"
)

const (
	DefaultDomain = "default-domain"
)

type InstanceMetadata struct {
	InstanceId string
	NicId      string
	MacAddress string
	IpAddress  string
	Gateway    string
}

type NetworkManager interface {
	Build(tenant, network, instanceName string) (*InstanceMetadata, error)
}

type NetworkManagerImpl struct {
	client        contrail.ApiClient
	privateSubnet string
	allocator     AddressAllocator
	instanceMgr   InstanceManager
}

func NewNetworkManager(server string, port int, privateSubnet string) NetworkManager {
	manager := new(NetworkManagerImpl)
	manager.client = contrail.NewClient(server, port)
	manager.privateSubnet = privateSubnet
	manager.allocator = NewAddressAllocator(manager.client, privateSubnet)
	manager.instanceMgr = NewInstanceManager(manager.client, manager.allocator)
	return manager
}

func (m *NetworkManagerImpl) Build(tenant, networkName, instanceName string) (*InstanceMetadata, error) {
	network := m.LocateNetwork(tenant, networkName)
	if network == nil {
		return nil, fmt.Errorf("Unable to loopkup or create network %s", networkName)
	}
	instance := m.instanceMgr.LocateInstance(tenant, instanceName)
	if instance == nil {
		return nil, fmt.Errorf("Unable to lookup or create instance %s", instanceName)
	}
	nic := m.instanceMgr.LocateInterface(network, instance)
	if nic == nil {
		return nil, fmt.Errorf("Unable to lookup or create interface for instance %s", instanceName)
	}
	ip := m.instanceMgr.LocateInstanceIp(network, nic)
	if ip == nil {
		return nil, fmt.Errorf("Unable to lookup or create instance-ip for instance %s", instanceName)
	}
	refs, err := network.GetNetworkIpamRefs()
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve network-ipam refs")
	}
	attr := refs[0].Attr.(types.VnSubnetsType)
	gateway := attr.IpamSubnets[0].DefaultGateway
	mdata := &InstanceMetadata{
		InstanceId: instance.GetUuid(),
		NicId:      nic.GetUuid(),
		MacAddress: nic.GetVirtualMachineInterfaceMacAddresses().MacAddress[0],
		IpAddress:  ip.GetInstanceIpAddress(),
		Gateway:    gateway,
	}
	return mdata, nil
}

func (m *NetworkManagerImpl) LocateNetwork(tenant, networkName string) *types.VirtualNetwork {
	fqn := []string{DefaultDomain, tenant, networkName}
	obj, err := m.client.FindByName("virtual-network", strings.Join(fqn, ":"))
	if err == nil {
		return obj.(*types.VirtualNetwork)
	}

	projectId, err := m.client.UuidByName("project", DefaultDomain+":"+tenant)
	if err != nil {
		glog.Infof("GET %s: %v", tenant, err)
		return nil
	}
	uid, err := config.CreateNetworkWithSubnet(
		m.client, projectId, networkName, m.privateSubnet)
	if err != nil {
		glog.Infof("Create %s: %v", networkName, err)
		return nil
	}
	obj, err = m.client.FindByUuid("virtual-network", uid)
	if err != nil {
		glog.Infof("GET %s: %v", networkName, err)
		return nil
	}
	glog.Infof("Create network %s", networkName)
	return obj.(*types.VirtualNetwork)
}
