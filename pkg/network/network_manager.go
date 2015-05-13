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
	network, err := m.LocateNetwork(tenant, networkName)
	log.Debug("Located Network: %s", network.GetDisplayName())
	if err != nil {
		return nil, fmt.Errorf("unable to loopkup or create network: %s", networkName, err)
	}

	instance, err := m.instanceMgr.LocateInstance(tenant, instanceName)
	log.Debug("Located Instance: %s", instance.GetDisplayName())
	if err != nil {
		return nil, fmt.Errorf("unable to lookup or create instance %s: %s", instanceName, err)
	}

	nic, err := m.instanceMgr.LocateInterface(network, instance)
	log.Debug("Located NIC: %s", nic.GetDisplayName())
	if err != nil {
		return nil, fmt.Errorf("Unable to lookup or create interface for instance %s: %s", instanceName, err)
	}

	ip, err := m.instanceMgr.LocateInstanceIp(network, nic)
	log.Debug("Located IP: %s", ip.GetDisplayName())
	if err != nil {
		return nil, fmt.Errorf("Unable to lookup or create instance-ip for instance %s: %s", instanceName, err)
	}

	gateway, err := m.instanceMgr.LocateInstanceGateway(network)
	log.Debug("Located Gateway: %s", gateway)
	if err != nil {
		return nil, fmt.Errorf("Unable to get instance gateway: %s", err)
	}

	macAddress, err := m.instanceMgr.LocateMacAddress(strings.Join(instanceFQName(tenant, instanceName), ":"))
	log.Debug("Located MacAddress: %s", macAddress)
	if err != nil {
		return nil, fmt.Errorf("Unable to get instance mac address: %s", err)
	}

	mdata := &InstanceMetadata{
		InstanceId: instance.GetUuid(),
		NicId:      nic.GetUuid(),
		MacAddress: macAddress,
		IpAddress:  ip.GetInstanceIpAddress(),
		Gateway:    gateway,
	}
	return mdata, nil
}

func (m *NetworkManagerImpl) LocateNetwork(tenant, networkName string) (*types.VirtualNetwork, error) {
	fqn := []string{DefaultDomain, tenant, networkName}
	vn, err := types.VirtualNetworkByName(m.client, strings.Join(fqn, ":"))

	// If there is an error since it doesn't exist yet, create it.
	if err != nil && vn == nil {
		projectName := fmt.Sprintf("%s:%s", DefaultDomain, tenant)

		log.Debug("ProjectByName: %s", projectName)
		project, err := types.ProjectByName(m.client, projectName)
		if err != nil {
			log.Error("GET %s: %v", tenant, err)
			return nil, err
		}

		log.Debug("CreateNetworkWithSubnet: project_id=%s, name=%s, prefix=%s", project.GetDisplayName(), networkName, m.privateSubnet)
		uid, err := config.CreateNetworkWithSubnet(m.client, project.GetUuid(), networkName, m.privateSubnet)
		if err != nil {
			log.Error("Create %s: %v", networkName, err)
			return nil, err
		}

		log.Debug("VirtualNetworkByUuid: %s", uid)
		vn, err = types.VirtualNetworkByUuid(m.client, uid)
		if err != nil {
			log.Error("GET %s: %v", networkName, err)
			return nil, err
		}

		log.Info("Created network %s", networkName)
	}

	return vn, nil
}
