package main

// Example of incoming YAML data structure for test purposes
var ExampleYamlData = []byte(
	`
- name: Test
  vars:
    insights_signature: "ascii_armored gpg signature"
    insights_signature_exclude: "/vars/insights_signature,/vars/content_vars"
    interpreter: /bin/bash
    content: |
        #!/bin/sh
        echo "$RHC_WORKER_FOO $RHC_WORKER_BAR!"
    content_vars:
        FOO: Hello
        BAR: World`)
