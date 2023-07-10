![Tests](https://github.com/oamg/rhc-worker-bash/actions/workflows/tests.yml/badge.svg)
[![codecov](https://codecov.io/github/oamg/rhc-worker-bash/branch/main/graph/badge.svg?token=6MRLOJS2SJ)](https://codecov.io/github/oamg/rhc-worker-bash)


# RHC Worker Bash

Remote Host Configuration (rhc) worker for executing bash scripts on hosts
managed by Red Hat Insights.

## Contact
* Package maintainer: Rodolfo Olivieri - rolivier@redhat.com

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

* Script to be executed and data host serving our script
    * Example is present inside the folder `development/nginx`
    * **Set it up yourself**  - see [Bash script example](#bash-script-example) below
* System connected via rhc with running rhcd == the system on which the script will be executed
    * **Set it up yourself** - for vagrant box see commands below
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
* MQTT broker for sending messages
    * Set up as part of `make development` call
* Storage to simulate upload by the ingress service
    * Set up as part of `make development` call

### Publish first message

1. Have system connected with rhc and running rhcd
    * depends on you if you want to use vagrant or different approach
2. Start MQTT broker, data host for serving the script and minio storage
    * You can take advantage of `make development` command to create neccessary containers, inspect `development/podman-compose.yml` for more details
3. Publish new message to broker
    * Change `CLIENT_ID` on L16 in `development/python/mqtt_publish.py` file if needed
    * Call `make publish`
4. You should see logs in rhcd and file with stdout of your script uploaded to the minio storage
    * Go to http://localhost:9990/login and use credentials from `.env` file

### Bash script example

*NOTE: This is subject to changes, right now worker is executing raw bash script, but in near future we expect that worker will execute bash script wrapped in signed yaml file.*

Create a bash script file called `command` inside the folder `development/nginx` and place
your code inside of it, like for example:

```bash
/usr/bin/convert2rhel --help
```
