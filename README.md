![Tests](https://github.com/oamg/rhc-worker-script/actions/workflows/tests.yml/badge.svg)
[![codecov](https://codecov.io/github/oamg/rhc-worker-script/branch/main/graph/badge.svg?token=6MRLOJS2SJ)](https://codecov.io/github/oamg/rhc-worker-script)

# RHC Worker

Remote Host Configuration (rhc) worker for executing  scripts on hosts
managed by Red Hat Insights. Interpreter used to execute the script is defined inside the supplied yaml file - served by insights.

- [RHC Worker Bash](#rhc-worker-script)
  - [General workflow of the worker](#general-workflow-of-the-worker)
  - [Getting started with local development](#getting-started-with-local-development)
    - [Publish first message](#publish-first-message)
    - [Worker playbooks](#worker-playbooks)
      - [Custom playbook](#custom-playbook)
      - [Convert2RHEL Playbook](#convert2rhel-playbook)
  - [FAQ](#faq)
    - [Are there special environment variables used by `rhc-worker-script`?](#are-there-special-environment-variables-used-by-rhc-worker-script)
    - [Can I change behavior of `rhc-worker-script`?](#can-i-change-behavior-of-rhc-worker-script)
    - [Can I change the location of `rhc-worker-script` config?](#can-i-change-the-location-of-rhc-worker-script-config)
  - [Contact](#contact)
    - [Package maintainers](#package-maintainers)

## General workflow of the worker

Everything starts when message is sent to rhcd. Worker then:

1. Picks up the message from rhcd
2. Downloads the worker playbook as temporary file (see [Worker playbooks](#worker-playbooks))
3. Verify the integrity of the playbook with `insights-client`
4. Executes the script
5. Reads stdout of the script
6. Sends the stdout wrapped in JSON back to rhcd

Then rhcd sends the message to upload service (with data from worker) in order to show the results in Insights UI - our setup for local development simulates the upload with minio storage.

## Getting started with local development

Almost everything that is needed for local development is placed in `development` folder.

Overview of what is needed:

- Script to be executed and data host serving our script
  - Example is present inside the folder `development/nginx/data`
  - **Set it up yourself**  - see [Worker playbooks](#worker-playbooks) below
- System connected via rhc with running rhcd == the system on which the script will be executed
  - **Set it up yourself** - for vagrant box see commands below

```bash
# Get a new centos-7 box
vagrant init eurolinux-vagrant/centos-7

# Install insights-client and rhc
...

# Connect via rhc
vagrant ssh -- -t 'rhc connect --server=$(RHSM_SERVER_URL) --username=$(RHSM_USERNAME) --password=$(RHSM_PASSWORD)'
# Run rhcd
vagrant ssh -- -t 'rhcd --log-level trace \
    --socket-addr $(YGG_SOCKET_ADDR) \
    --broker $(HOST_IP):1883 \
    --topic-prefix yggdrasil \
    --data-host $(HOST_IP):8000'
```

- MQTT broker for sending messages
  - Set up as part of `make development` call
- Storage to simulate upload by the ingress service
  - Set up as part of `make development` call

### Publish first message

1. Have system connected with rhc and running rhcd
    - depends on you if you want to use vagrant or different approach
2. Start MQTT broker, data host for serving the script and minio storage
    - You can take advantage of `make development` command to create neccessary containers, inspect `development/podman-compose.yml` for more details
3. Publish new message to broker
    - [Optional] Change values in `development/python/mqtt_publish.py`
      - `CLIENT_ID` - can be found in logs after running rhcd
      - `SERVED_FILENAME` - one of the files inside `development/nginx/data`
    - Call `make publish`
4. You should see logs in rhcd and file with stdout of your script uploaded to the minio storage
    - Go to <http://localhost:9990/login> and use credentials from `.env` file

### Worker playbooks

There is an [example bash playbook](
https://github.com/oamg/rhc-worker-script/blob/main/development/nginx/data/example_bash.yaml)
available under `development/nginx/data`, with a minimal bash script to use
during the worker execution.

If there's a need to test any other playbook provided in this repository, one
must change what playbook will be used during the message consumption in the
[mqtt_publish.py](https://github.com/oamg/rhc-worker-script/blob/main/development/python/mqtt_publish.py#L22)
file with the name that corresponds the ones present in `development/nginx/data`. Currently, the ones available are:

1. [example_bash.yaml](https://github.com/oamg/rhc-worker-script/blob/main/development/nginx/data/example_bash.yaml)
2. [example_python.yaml](https://github.com/oamg/rhc-worker-script/blob/main/development/nginx/data/example_python.yaml)
3. [convert2rhel.yaml](https://github.com/oamg/rhc-worker-script/blob/main/development/nginx/data/convert2rhel.yaml)

#### Custom playbook

Create or update a yaml file inside the folder `development/nginx/data/*`.
Correct structure with exampe bash script can be seen below:

#### Convert2RHEL Playbook

A specialized [Convert2RHEL](https://github.com/oamg/convert2rhel) playbook can be found under the `development/nginx/data` as well. The playbook will take of the following functions:

1. Setup Convert2RHEL (Download certificates, repositories and etc...)
2. Set a couple of environment variables for the Convert2RHEl execution (Based on the `content_vars` defined in the playbook)
3. Run convert2rhel with default commands
4. A function to run any post-execution commands needed by the conversion (Currently empty.)

## FAQ

### Are there special environment variables used by `rhc-worker-script`?

There is one special variable that must be set in order to run our worker and that is `YGG_SOCKET_ADDR`, this variable value is set by `rhcd` via `--socket-addr` option.

Other than that there are no special variables, however if downloaded yaml file contained `content_vars` (like the example above), then before the execution of the bash script (`content`) all such variables are set as environment variables and prefixed with `RHC_WORKER_`, after script execution is done they are unset.

### Can I change behavior of `rhc-worker-script`?

Yes, some values can be changed if config exists at `/etc/rhc/workers/rhc-worker-script.yml`, **the config must have valid yaml format**, see all available fields below.

Example of full config (with default values):

```yaml
# rhc-worker-script configuration

# recipient directive to register with dispatcher
directive: "rhc-worker-script"

# whether to verify incoming yaml files
verify_yaml: true

# perform the insights-client GPG check on the insights-core egg
insights_core_gpg_check: true

# temporary directory in which the temporary files with executed bash scripts are created
temporary_worker_directory: "/var/lib/rhc-worker-script"
```

### Can I change the location of `rhc-worker-script` config?

No, not right now. If you want this feature please create an issue or upvote already existing issue.

## Contact

### Package maintainers

- Rodolfo Olivieri - <rolivier@redhat.com>
- Andrea Waltlova - <awaltlov@redhat.com>
