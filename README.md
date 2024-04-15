# OS-diff
OpenStack / OpenShift diff tool

This tool collects OpenStack/OpenShift service configurations,
compares configuration files, makes a diff and creates a report to the user
in order to provide information and warnings after a migration from
OpenStack to OpenStack on OpenShift migration.

### Usage

#### Setup our environment

##### Setup Ssh access

In order to allow Os-diff to connect to our clouds and pull files from the services you describe in the `config.yaml`file you need to properly set the option in the `os-diff.cfg`:

```
[Default]

local_config_dir=/tmp/
service_config_file=config.yaml

[Tripleo]

ssh_cmd=ssh -F ssh.config
director_host=standalone
container_engine=podman
connection=ssh
remote_config_path=/tmp/tripleo
local_config_path=/tmp/

[Openshift]

ocp_local_config_path=/tmp/ocp
connection=local
ssh_cmd=""

```

The `ssh_cmd` will be used by os-diff to access via ssh to your TripleO Undercloud/Director host or the host where your cloud is accessible and the podman/docker binary is installed and allowed to interract with the running containers.
This option could have different form:

```
ssh_cmd=ssh -F ssh.config standalone
director_host=
```

```
ssh_cmd=ssh -F ssh.config
director_host=standalone
```

or without a ssh config file:

```
ssh_cmd=ssh -i /home/user/.ssh/id_rsa stack@my.undercloud.local
director_host=
```

or

```
ssh_cmd=ssh -i /home/user/.ssh/id_rsa stack@
director_host=my.undercloud.local
```

Note that the concat of ssh_cmd + director_host should be a "successful ssh access".

##### Generate ssh.config file from inventory or hosts file

Os-diff can use an ssh.config file for getting access to your TripleO/OSP environment.
A command can help you to generate this ssh config file from your Ansible inventory (like tripleo-ansible-inventory.yaml file):

```
os-diff configure -i tripleo-ansible-inventory.yaml -o ssh.config --yaml
```

The ssh.config file will looks like this (for a Standalone deployment):

```
Host standalone
  HostName standalone
  User root
  StrictHostKeyChecking no
  UserKnownHostsFile /dev/null

Host undercloud
  HostName undercloud
  User root
  StrictHostKeyChecking no
  UserKnownHostsFile /dev/null
```

Note: You will have to set the IdentityFile key in the file in order to get a fully working acces:

```
Host standalone
  HostName standalone
  User root
  IdentityFile ~/.ssh/id_rsa
  StrictHostKeyChecking no
  UserKnownHostsFile /dev/null

Host undercloud
  HostName undercloud
  User root
  IdentityFile ~/.ssh/id_rsa
  StrictHostKeyChecking no
  UserKnownHostsFile /dev/null
```

#### Non-standard services settings

It's important to configure correctly a ssh config file or equivalent for non standard services such as OVS.
The ovs_external_ids is not a service which runs in a container and the ovs data are stored on each hosts of our cloud: controller_1/controller_2/...

The hosts key in the config.yaml will allow os-diff to loop and get for all hosts specified the output of the command or the file or the data that you need to pull from our deployment in order to compare it later:

```
    ovs_external_ids:
        path:
            - ovs_external_ids.json
        hosts:
            - standalone
```

The `service_command` is the command which provides the required informations. It could a simple cat from a config file.
`cat_output` should be set to true if you want os-diff to get the output of the command and stored the output in a file specified by the key `path`

Then you can provide a mapping between, in this case the EDPM CRD and the ovs-vsctl output with `config_mapping`

```
        service_command: 'ovs-vsctl list Open_vSwitch . | grep external_ids | awk -F '': '' ''{ print $2; }'''
        cat_output: true
        config_mapping:
            ovn-bridge: edpm_ovn_bridge
            ovn-bridge-mappings: edpm_ovn_bridge_mappings
            ovn-encap-type: edpm_ovn_encap_type
            ovn-match-northd-version: ovn_match_northd_version
            ovn-monitor-all: ovn_monitor_all
            ovn-ofctrl-wait-before-clear: edpm_ovn_ofctrl_wait_before_clear
            ovn-remote-probe-interval: edpm_ovn_remote_probe_interval
```
Then you can use this command to compare the values:

```
os-diff diff ovs_external_ids.json edpm.crd --crd --service ovs_external_ids
```

#### Pull configuration step

Before running the Pull command you need to configure the SSH access to your environments (OpenStack and OCP).
Edit os-diff.cfg and/or the ssh.config provided with this project and make sure you can ssh on your hosts 
without password or host key verification, with the command:

```
ssh -F ssh.config crc
ssh -F ssh.config standalone
```

When everything is setup correctly you can tweak the config.yaml file at the root of the project which will contain the description of the services you want to extract configuration from:

```
  config.yaml
```

You can add your own service(s) according to the following:

```
  # Service name
  keystone:
    # Bool to enable/disable a service (not implemented yet)
    enable: true
    # Pod name, in both OCP and podman context.
    # It could be strict match with strict_pod_name_match set to true
    # or by default it will just grep the podman and work with all the pods
    # which matched with pod_name.
    podman_name: keystone
    pod_name: keystone
    container_name: keystone-api
    # Path of the config files you want to analyze.
    # It could be whatever path you want:
    # /etc/<service_name> or /etc or /usr/share/<something> or even /
    # @TODO: need to implement loop over path to support multiple paths such as:
    # - /etc
    # - /usr/share
    path:
      - /etc/
      - /etc/keystone
      - /etc/keystone/keystone.conf
      - /etc/keystone/logging.conf
```

Once everything is correctly setup you can start to pull configuration:

```
# build os-diff
make build
# run pull configuration for TripleO standalone:
./os-diff pull --env=tripleo
# run pull configuration for OCP with a specific output directory and a specific service config file:
./os-diff pull -e ocp -o /tmp/myconfigdir -s my-service-config-file
```

Note: The CLI arguments take precedence on the configuration file values.

#### Compare configuration files steps

Once you have collected all the data per services you need, you can start to run comparison between
your two source directories.
A results file is written at the root of this project `results.log` and a *.diff file is created for each
file where a difference has been detected

```diff
/tmp/collect_crc_configs/nova/nova-api-0/etc/nova/nova.conf.diff

# with this kind of content:
Source file path: /tmp/collect_crc_configs/nova/nova-api-0/etc/nova/nova.conf, difference with: /tmp/collect_crc_configs/nova/nova-cell0-conductor-0/etc/nova/nova.conf
[DEFAULT]
-transport_url=rabbit://default_user_pVPGFkYMWTdSarUSog9:Rg59ofmjeDWg24v8ZeGW-1PblH1LJDQ1@rabbitmq.openstack.svc:5672
[api]
-auth_strategy=keystone
```

The log INFO/WARN and ERROR will be print to the console as well so you can have colored info regarding the current file processing.
Run the compare command:

```
./os-diff diff /tmp/collect_tripleo_configs /tmp/collect_crc_configs

```

### Examples:

diff command compares file to file only and ouput a diff with color on the console.
Example for YAML file:

```diff
./os-diff diff tests/podman/key.yaml tests/ocp/key.yaml
Source file path: tests/podman/key.yaml, difference with: tests/ocp/key.yaml
@@ line: 8
+    pod_name: foo
@@ line: 2
-    pod_name: keystone
```

Example for ini config file:

```diff
 ./os-diff diff /tmp/collect_ocp_configs/keystone/etc/keystone/keystone.conf /tmp/collect_tripleo_configs/keystone/etc/keystone/keystone.conf
Source file path: /tmp/collect_ocp_configs/keystone/etc/keystone/keystone.conf, difference with: /tmp/collect_tripleo_configs/keystone/etc/keystone/keystone.conf
[DEFAULT]
-use_stderr=true
-notification_format=basic
-debug=True
-transport_url=rabbit://guest:xM2nhUiV60xoEPTjoxNJ5vFWC@undercloud-0.ctlplane.redhat.local:5672/?ssl=0
[cache]
-backend=dogpile.cache.memcached
-enabled=True
-memcache_servers=undercloud-0.ctlplane.redhat.local:11211
-tls_enabled=False
[catalog]
-driver=sql
[cors]
-allowed_origin=*
[credential]
-key_repository=/etc/keystone/credential-keys
[database]
+connection=mysql+pymysql://keystone:12345678@openstack/keystone
-connection=mysql+pymysql://keystone:xxx@192.168.24.3/keystone?read_default_file=/etc/my.cnf.d/tripleo.cnf&read_default_group=tripleo
[fernet_tokens]
+max_active_keys=2
-max_active_keys=5
```

### OpenShift Pod config comparison

When you prepare the adoption of your TripleO cloud to your OpenShift cluster, you might want to compare and verify if the config described in your OpenShift config desc file has no difference with your Tripleo service config or even, want to verify that after patching the OpenShift config, the service is correctly configured.

The service command allow you to compare YAML OpenShift config patch with OpenStack Ini configuration file from your services.
You can also query OpenShift pods to check if the configuration are correct.

Example:

```
spec:
  glance:
    enabled: true
    template:
      databaseInstance: openstack
      containerImage: foo
      customServiceConfig: |
        [DEFAULT]
        enabled_backends=default_backend:rbd
        [glance_store]
        default_backend=default_backend
        [default_backend]
        rbd_store_ceph_conf=/etc/ceph/ceph.conf
        rbd_store_user=openstack
        rbd_store_pool=images
        store_description=Ceph glance store backend.
...
```

Run service command:

```service
./os-diff cdiff -s glance -o examples/glance/glance.patch -c /tmp/glance.conf
Source file path: examples/glance/glance.patch, difference with: /tmp/glance.conf
-enabled_backends=default_backend:rbd
-[glance_store]
-default_backend=default_backend
-[default_backend]
-rbd_store_ceph_conf=/etc/ceph/ceph.conf
-rbd_store_user=openstack
-rbd_store_pool=images
-store_description=Ceph glance store backend.
```

Run comparison against the deployed pod:

```service
./os-diff cdiff -s glance -o examples/glance/glance.patch -c /etc/glance/glance-api.conf \
--frompod -p glance-external-api-678c6c79d7-24t7t

Source file path: examples/glance/glance.patch, difference with: /etc/glance/glance-api.conf
[DEFAULT]
-enabled_backends=default_backend:rbd
[glance_store]
-default_backend=default_backend
-[default_backend]
-rbd_store_ceph_conf=/etc/ceph/ceph.conf
-rbd_store_user=openstack
-rbd_store_pool=images
-store_description=Ceph glance store backend.
```

### Add service

If you want to add a new OpenStack service to this tool follow those instructions:

* Convert your OpenShift configmap to a GO struct with:
https://zhwt.github.io/yaml-to-go/
* Create a <service-name>.go file into pkg/servicecfg/
* Paste your generated structure and the following code:
```
package servicecfg

import (
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

type YourServiceName struct {
	Spec struct {
		YourServiceName struct {
      Template: {
        CustomServiceConfig string `yaml:"customServiceConfig"`
      }
    }
  }
}
```
### Asciinema demo

https://asciinema.org/a/618124

### TODO

* Improve reporting (console, debug and log file with general report)
* Improve diff output for json and yaml
* Improve Makefile entry with for example: make compare
* Add a skip list (skip /etc/keystone/fernet-keys )
* Add interactive and edit mode to ask for editing the config for the user
  when a difference has been found

