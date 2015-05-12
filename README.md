# packnet

Command for connecting Docker containers to OpenContrail

## Step 1. Start a packnet container

Note:

* **The containers must run in privileged mode with some special flags.**

```
[root@computenode001 ~]# docker run -it --rm --privileged --net=host --pid=host -v /var/run/docker.sock:/var/run/docker.sock dockers.tf.riotgames.com/rcluster/packnet /bin/bash
```

## Step 2. Start base container

```
[root@computenode001 ~]# docker run -d --name steve_test --net=none dockers.tf.riotgames.com/rcluster/base
```

## Step 3. Connect container to OpenContrail in packnet container

```
app$ ./packnet --network=globalqa.pdx2.steve.test --server=10.142.208.9 --tenant=steve.test --start=<container-id>
```

### Step 4. Use the network settings from the container in additional containers

```
[root@computenode001 ~]# docker run --rm -it --net="container:steve_test" cirros /bin/sh
```