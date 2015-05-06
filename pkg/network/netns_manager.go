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
	"os/exec"
	"strconv"

	"github.com/docker/libcontainer/netlink"
	"github.com/milosgajdos83/tenus"
)

type NetnsManager interface {
	CreateInterface(dockerId, macAddress, ipAddress, gateway string) (string, error)
	DeleteInterface(dockerId string) error
}

type NetnsManagerImpl struct {
}

func NewNetnsManager() NetnsManager {
	m := new(NetnsManagerImpl)
	return m
}

func (m *NetnsManagerImpl) CreateInterface(dockerId, macAddress, ipAddress, gateway string) (string, error) {
	masterName := fmt.Sprintf("veth-%s", dockerId[0:10])
	veth, err := tenus.NewVethPairWithOptions(masterName, tenus.VethOptions{PeerName: "veth0"})
	if err != nil {
		return "", err
	}
	pid, err := tenus.DockerPidByName(dockerId, "/var/run/docker.sock")
	if err != nil {
		return "", err
	}
	veth.SetPeerLinkNsPid(pid)
	peer := veth.PeerNetInterface()
	netlink.NetworkSetMacAddress(peer, macAddress)
	veth.SetLinkUp()

	cmd := exec.Command("nsenter", "-n", "-t", strconv.Itoa(pid),
		"ip", "link", "set", "veth0", "up")
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	cmd = exec.Command("nsenter", "-n", "-t", strconv.Itoa(pid), "ip", "addr", "add",
		fmt.Sprintf("%s/32", ipAddress), "peer", gateway, "dev", "veth0")
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	cmd = exec.Command("nsenter", "-n", "-t", strconv.Itoa(pid), "ip", "route", "add",
		"default", "via", gateway)
	err = cmd.Run()

	return masterName, nil
}

func (m *NetnsManagerImpl) DeleteInterface(dockerId string) error {
	// masterName := fmt.Sprintf("veth-%s", dockerId[0:10])
	return nil
}
