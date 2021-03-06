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

package main

import (
	"os"
	"os/exec"

	"github.com/op/go-logging"
	flag "github.com/spf13/pflag"

	"github.com/pedro-r-marques/packnet/pkg/network"
)

var log = logging.MustGetLogger("packnet")

type Config struct {
	ApiServer     string
	ApiPort       int
	Tenant        string
	NetworkName   string
	DockerId      string
	PrivateSubnet string
}

func init() {
	format := "%{module}[%{pid}]: %{time:2006-01-02T15:04:05Z} [%{shortfile}] %{level:s} - %{message}"
	logging.SetFormatter(logging.MustStringFormatter(format))
	logging.SetBackend(logging.NewLogBackend(os.Stderr, "", 0))
	logging.SetLevel(logging.DEBUG, "")
}

func main() {

	config := &Config{
		ApiServer:     "localhost",
		ApiPort:       8082,
		Tenant:        "teemo",
		NetworkName:   "default",
		PrivateSubnet: "10.40.128.0/17",
	}
	AddFlags(config, flag.CommandLine)
	flag.Parse()

	// Truncate Id to 11 digits for consistency
	config.DockerId = config.DockerId[0:10]

	if flag.Lookup("start").Value.String() != "" {
		Start(config)
	} else if flag.Lookup("stop").Value.String() != "" {
		Stop(config)
	}
}

func AddFlags(c *Config, fs *flag.FlagSet) {
	fs.StringVar(&c.ApiServer, "server", c.ApiServer, "OpenContrail API server.")
	fs.StringVar(&c.Tenant, "tenant", c.Tenant, "Administrative domain.")
	fs.StringVar(&c.NetworkName, "network", c.NetworkName, "Network identifier")
	fs.StringVar(&c.DockerId, "start", "", "Provision the network of the container")
	fs.StringVar(&c.DockerId, "stop", "", "Provision the network of the container")
}

func Start(c *Config) error {
	manager := network.NewNetworkManager(c.ApiServer, c.ApiPort, c.PrivateSubnet)
	metadata, err := manager.Build(c.Tenant, c.NetworkName, c.DockerId)
	if err != nil {
		log.Fatal(err)
	}
	nsMan := network.NewNetnsManager()
	masterName, err := nsMan.CreateInterface(c.DockerId, metadata.MacAddress, metadata.IpAddress, metadata.Gateway)
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}

	cmd := exec.Command("vrouter-ctl", "--mac-address", metadata.MacAddress,
		"--vm", metadata.InstanceId, "--vmi", metadata.NicId,
		"--interface", masterName, "add", c.DockerId)
	err = cmd.Run()
	if err != nil {
		out, _ := cmd.CombinedOutput()
		log.Fatal(err.Error() + ": " + string(out))
	}
	return nil
}

func Stop(c *Config) error {
	nsMan := network.NewNetnsManager()
	nsMan.DeleteInterface(c.DockerId)
	return nil
}
