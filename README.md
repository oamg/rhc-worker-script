![Tests](https://github.com/oamg/rhc-worker-bash/actions/workflows/tests.yml/badge.svg)
[![codecov](https://codecov.io/github/oamg/rhc-worker-bash/branch/main/graph/badge.svg?token=6MRLOJS2SJ)](https://codecov.io/github/oamg/rhc-worker-bash)

# RHC Worker Bash

Remote Host Configuration (rhc) worker for executing bash scripts on hosts
managed by Red Hat Insights.

- [RHC Worker Bash](#rhc-worker-bash)
  - [General workflow of the worker](#general-workflow-of-the-worker)
  - [Getting started with local development](#getting-started-with-local-development)
    - [Publish first message](#publish-first-message)
    - [Bash script example](#bash-script-example)
  - [FAQ](#faq)
    - [Are there special environment variables used by `rhc-worker-bash`?](#are-there-special-environment-variables-used-by-rhc-worker-bash)
    - [Can I change behavior of `rhc-worker-bash`?](#can-i-change-behavior-of-rhc-worker-bash)
    - [Can I change the location of `rhc-worker-bash` config?](#can-i-change-the-location-of-rhc-worker-bash-config)
  - [Contact](#contact)
    - [Package maintainers](#package-maintainers)

## General workflow of the worker

Everything starts when message is sent to rhcd. Worker then:

1. Picks up the message from rhcd
2. Downloads the bash script as temporary file (see [Bash script example](#bash-script-example))
3. Executes the script
4. Reads stdout of the script
5. Sends the stdout wrapped in JSON back to rhcd

Then rhcd sends the message to upload service (with data from worker) in order to show the results in Insights UI - our setup for local development simulates the upload with minio storage.

## Getting started with local development

Almost everything that is needed for local development is placed in `development` folder.

Overview of what is needed:

- Script to be executed and data host serving our script
  - Example is present inside the folder `development/nginx`
  - **Set it up yourself**  - see [Bash script example](#bash-script-example) below
- System connected via rhc with running rhcd == the system on which the script will be executed
  - **Set it up yourself** - for vagrant box see commands below

```bash
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

### Bash script example

Create or update a yaml file inside the folder `development/nginx/data/*`.
Correct structure with exampe bash script can be seen below:

```yml
vars:
  _insights_signature: |
    ascii_armored gpg signature
  _insights_signature_exclude: "/vars/insights_signature,/vars/content_vars"
  content: |
    #!/bin/sh
    /usr/bin/convert2rhel --help
  content_vars:
    # variables that will be handed to the script as environment vars
    # will be prefixed with RHC_WORKER_*
    FOO: bar
    BAR: foo
```

## FAQ

### Are there special environment variables used by `rhc-worker-bash`?

There is one special variable that must be set in order to run our worker and that is `YGG_SOCKET_ADDR`, this variable value is set by `rhcd` via `--socket-addr` option.

Other than that there are no special variables, however if downloaded yaml file contained `content_vars` (like the example above), then before the execution of the bash script (`content`) all such variables are set as environment variables and prefixed with `RHC_WORKER_`, after script execution is done they are unset.

### Can I change behavior of `rhc-worker-bash`?

Yes, some values can be changed if config exists at `/etc/rhc/workers/rhc-worker-bash.yml`, **the config must have valid yaml format**, see all available fields below.

Example of full config (with default values):

```yaml
# rhc-worker-bash configuration

# recipient directive to register with dispatcher
directive: "rhc-worker-bash"

# whether to verify incoming yaml files
verify_yaml: true

# perform the insights-client GPG check on the insights-core egg
insights_core_gpg_check: true

# temporary directory in which the temporary files with executed bash scripts are created
temporary_worker_directory: "/var/lib/rhc-worker-bash"
```

### Can I change the location of `rhc-worker-bash` config?

No, not right now. If you want this feature please create an issue or upvote already existing issue.

## Contact

### Package maintainers

- Rodolfo Olivieri - <rolivier@redhat.com>
- Andrea Waltlova - <awaltlov@redhat.com>
