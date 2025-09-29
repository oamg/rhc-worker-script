![Tests](https://github.com/oamg/rhc-worker-script/actions/workflows/verify.yml/badge.svg)
[![pre-commit.ci status](https://results.pre-commit.ci/badge/github/oamg/rhc-worker-script/main.svg)](https://results.pre-commit.ci/latest/github/oamg/rhc-worker-script/main)
[![codecov](https://codecov.io/github/oamg/rhc-worker-script/branch/main/graph/badge.svg?token=6MRLOJS2SJ)](https://codecov.io/github/oamg/rhc-worker-script)

# RHC Worker

Remote Host Configuration (rhc) worker for executing  scripts on hosts managed
by Red Hat Lightspeed. Interpreter used to execute the script is defined inside
the supplied yaml file - served by insights.

- [RHC Worker](#rhc-worker)
  - [General workflow of the worker](#general-workflow-of-the-worker)
  - [FAQ](#faq)
    - [Are there special environment variables used by `rhc-worker-script`?](#are-there-special-environment-variables-used-by-rhc-worker-script)
    - [Can I change behavior of `rhc-worker-script`?](#can-i-change-behavior-of-rhc-worker-script)
    - [Can I change the location of `rhc-worker-script` config?](#can-i-change-the-location-of-rhc-worker-script-config)
  - [Contact](#contact)
    - [Package maintainers](#package-maintainers)

## General workflow of the worker

Everything starts when message is sent to rhcd. Worker then:

1. Picks up the message from rhcd
2. Downloads the worker playbook as temporary file (see [Worker playbooks](https://github.com/oamg/convert2rhel-insights-tasks/blob/main/playbooks/))
3. Verify the integrity of the playbook with `insights-client`
4. Executes the script
5. Reads stdout of the script
6. Sends the stdout wrapped in JSON back to rhcd

Then rhcd sends the message to upload service (with data from worker) in order
to show the results in Red Hat Lightspeed UI - our setup for local development
simulates the upload with minio storage.

## FAQ

### Are there special environment variables used by `rhc-worker-script`?

There is one special variable that must be set in order to run our worker and that is `YGG_SOCKET_ADDR`, this variable value is set by `rhcd` via `--socket-addr` option.

Other than that there are no special variables, however if downloaded yaml file contained `content_vars` (like the example above), then before the execution of the bash script (`content`) all such variables are set as environment variables and prefixed with `RHC_WORKER_`, after script execution is done they are unset.

### Can I change behavior of `rhc-worker-script`?

Yes, some values can be changed in the config file located at `/etc/rhc/workers/rhc-worker-script.yml`. After installing the `rhc-worker-script` package, a config file will be created with the default values required for the worker to start processing messages, **the config must have valid yaml format**, see all available fields in the [rhc-worker-script.yml](https://github.com/oamg/rhc-worker-script/blob/main/rhc-worker-script.yml) in the root of the repository.

### Can I change the location of `rhc-worker-script` config?

No, not right now. If you want this feature please create an issue or upvote already existing issue.

## Contact

### Package maintainers

- Rodolfo Olivieri - <rolivier@redhat.com>
- Andrea Waltlova - <awaltlov@redhat.com>
