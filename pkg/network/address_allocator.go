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
	"strings"

	"github.com/Juniper/contrail-go-api"
	"github.com/Juniper/contrail-go-api/config"
	"github.com/Juniper/contrail-go-api/types"
)

type AddressAllocator interface {
	LocateIpAddress(uid string) (string, error)
	ReleaseIpAddress(uid string)
}

// Allocate an unique address for each Pod.
type AddressAllocatorImpl struct {
	client        contrail.ApiClient
	network       *types.VirtualNetwork
	privateSubnet string
}

const (
	AddressAllocationNetwork = "default-domain:default-project:addr-alloc"
)

func NewAddressAllocator(client contrail.ApiClient, privateSubnet string) AddressAllocator {
	a := &AddressAllocatorImpl{
		client:        client,
		privateSubnet: privateSubnet,
	}

	a.network = a.initializeAllocatorNetwork()

	return a
}

func (a *AddressAllocatorImpl) initializeAllocatorNetwork() *types.VirtualNetwork {
	vn, err := types.VirtualNetworkByName(a.client, AddressAllocationNetwork)
	if err == nil {
		return vn
	}

	fqn := strings.Split(AddressAllocationNetwork, ":")
	parent := strings.Join(fqn[0:len(fqn)-1], ":")
	projectId, err := a.client.UuidByName("project", parent)
	if err != nil {
		log.Fatalf("%s: %v", parent, err)
	}

	netId, err := config.CreateNetworkWithSubnet(a.client, projectId, fqn[len(fqn)-1], a.privateSubnet)
	if err != nil {
		log.Fatalf("%s: %v", parent, err)
	}
	log.Info("Created network %s", AddressAllocationNetwork)
	network, err := types.VirtualNetworkByUuid(a.client, netId)
	if err != nil {
		log.Fatalf("Get virtual-network %s: %v", netId, err)
	}
	return network
}

func (a *AddressAllocatorImpl) allocateIpAddress(uid string) (contrail.IObject, error) {
	ipObj := new(types.InstanceIp)
	ipObj.SetName(uid)
	ipObj.AddVirtualNetwork(a.network)
	err := a.client.Create(ipObj)
	if err != nil {
		log.Error("Create InstanceIp %s: %v", uid, err)
		return nil, err
	}
	obj, err := types.InstanceIpByUuid(a.client, ipObj.GetUuid())
	if err != nil {
		log.Error("Get InstanceIp %s: %v", uid, err)
		return nil, err
	}
	return obj, err
}

func (a *AddressAllocatorImpl) LocateIpAddress(uid string) (string, error) {
	obj, err := a.client.FindByName("instance-ip", uid)
	if err != nil {
		obj, err = a.allocateIpAddress(uid)
		if err != nil {
			return "", err
		}
	}

	ipObj := obj.(*types.InstanceIp)
	return ipObj.GetInstanceIpAddress(), nil
}

func (a *AddressAllocatorImpl) ReleaseIpAddress(uid string) {
	objid, err := a.client.UuidByName("instance-ip", uid)
	if err != nil {
		err = a.client.DeleteByUuid("instance-ip", objid)
		if err != nil {
			log.Warning("Delete instance-ip: %v", err)
		}
	}
}
