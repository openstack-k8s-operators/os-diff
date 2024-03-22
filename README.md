# OS-diff
OpenStack / OpenShift diff tool

This tool collects OpenStack/OpenShift service configurations,
compares configuration files, makes a diff and creates a report to the user
in order to provide informations and warnings after a migration from
OpenStack to OpenStack on OpenShift migration.

### Usage

#### Pull configuration step

Before running the Pull command you need to configure the ssh access to your environements (OpenStack and OCP).
Edit os-diff.cfg and/or the ssh.config provided with this project and make sure you can ssh on your hosts 
without password or host key verification, with the command:

```
ssh -F ssh.config crc
ssh -F ssh.config standalone
```

When everything is setup correctly you can tweak the config.yaml file at the root of the project which contain the description
of the services you want to extract configurations:

```
  config.yaml
```

You can add your own service according to the following:

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

diff command compare file to file only and ouput a diff with color on the console.
Example for Yaml file:

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

When you prepare the adoption of your TripleO cloud to your OpenShift cluster, you might want to compare and verify if the config describe in your OpenShift config desc file has no difference with your Tripleo service config or even, want to verify that after patching the OpenShift config, the service is correctly configured.

The service command allow you to compare Yaml OpenShift config patch with OpenStack Ini configuration file from your services.
You can also query OpenShift pods to check if the configuration are well set.

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

