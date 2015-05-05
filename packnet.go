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
	flag "github.com/spf13/pflag"

	"github.com/pedro-r-marques/packnet/pkg/network"
)

type Config struct {
	ApiServer     string
	Tenant        string
	NetworkName   string
	DockerId      string
	PrivateSubnet string
}

func AddFlags(c *Config, fs *flag.FlagSet) {
	fs.StringVar(&c.ApiServer, "server", c.ApiServer,
		"OpenContrail API server.")
	fs.StringVar(&c.Tenant, "tenant", c.Tenant,
		"Administrative domain.")
	fs.StringVar(&c.NetworkName, "network", c.NetworkName,
		"Network identifier")
	fs.StringVar(&c.DockerId, "start", "",
		"Provision the network of the container")
}

func Start(c *Config) error {
	manager := network.NewNetworkManager(c.ApiServer, 8082, c.PrivateSubnet)
	manager.Build(c.Tenant, c.NetworkName, c.DockerId)
	return nil
}

func main() {
	config := &Config{
		ApiServer:     "localhost",
		Tenant:        "teeno",
		NetworkName:   "default",
		PrivateSubnet: "10.40.128.0/17",
	}
	AddFlags(config, flag.CommandLine)
	flag.Parse()
	if flag.Lookup("start").Value.String() != "" {
		Start(config)
	}
}
