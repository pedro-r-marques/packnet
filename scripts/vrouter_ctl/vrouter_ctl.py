import argparse
from contrail_vrouter_api.vrouter_api import ContrailVRouterApi

def main():
	parser = argparse.ArgumentParser()
	# "--mac-address", metadata.MacAddress,
	# 	"--vm", metadata.InstanceId, "--vmi", metadata.NicId,
	# 	"--interface", masterName, "add", c.DockerId)
	parser.add_argument('--mac-address')
	parser.add_argument('--vm')
	parser.add_argument('--vmi')
	parser.add_argument('--interface')
	parser.add_argument('command', choices=['add'])
	parser.add_argument('dockerId')
	args = parser.parse_args()
	if args.command == 'add':
		api = ContrailVRouterApi()
    	api.add_port(args.vm, args.vmi, args.interface, args.mac_address, port_type='NovaVMPort', display_name=args.dockerId)
	else:
    		print "No command specified"


if __name__ == "__main__":
	main()
