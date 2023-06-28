# RHC Worker Bash

Remote Host Configuration (rhc) worker for executing bash scripts on hosts
managed by Red Hat Insights.

## Getting started

Create a bash script file called `command` inside the folder `python` and place
your code inside of it, like for example:

```bash
/usr/bin/convert2rhel --help
```

After that, we are ready to publish a new message to the MQTT broker with:

```bash
make publish
```
