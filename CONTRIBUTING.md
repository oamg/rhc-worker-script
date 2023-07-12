# Contributing to RHC Worker Bash

The following is a set of guidelines for contributing to RHC Worker Bash codebase,
which are hosted in the [OAMG Organization](https://github.com/oamg) on GitHub.
These are mostly guidelines, not rules.

## What should I know before I get started?

Below are a list of things to keep in mind when developing and submitting
contributions to this repository.

1. All golang code must be compatible with versions 1.20.
2. The code should follow linting from golangci-lint.
3. All commits should have passed the pre-commit checks.
4. Don't change code that is not related to your issue/ticket, open a new
   issue/ticket if that's the case.

### Working with GitHub

If you are not sure on how GitHub works, you can read the quickstart guide from
GitHub to introduce you on how to get started at the platform. [GitHub
Quickstart - Hello
World](https://docs.github.com/en/get-started/quickstart/hello-world).

### Setting up Git

If you never used `git` before, GitHub has a nice quickstart on how to set it
up and get things ready. [GitHub Quickstart - Set up
Git](https://docs.github.com/en/get-started/quickstart/set-up-git)

### Forking a repository

Forking is necessary if you want to contribute, but if you
are unsure on how this work (Or what a fork is), head out to this quickstart
guide from GitHub. [GitHub Quickstart - Fork a
repo](https://docs.github.com/en/get-started/quickstart/fork-a-repo)

As an additional material, check out this Red Hat blog post about [What is an
open source
upstream?](https://www.redhat.com/en/blog/what-open-source-upstream)

### Collaborating with Pull Requests

Check out this guide from GitHub on how to collaborate with pull requests. This
is an in-depth guide on everything you need to know about PRs, forks,
contributions and much more. [GitHub - Collaborating with pull
requests](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests)

## Getting started with development

See README section [Getting started with local development](/README.md#getting-started-with-local-development)

### Dependencies for local development

We have some required dependencies you should have installed on your system
(either your local computer or a container) to get ready to write some
code:

Required dependencies:

- virtualenv
- python
- go
- pre-commit
- git
- podman

Optional dependencies:

- [make](https://www.gnu.org/software/make/#download)

### Setting up the environment

The commands below will create a python3 virtual environment with all the
necessary dependencies installed and setup `pre-commit` hooks.

Beware this command can take a while to finish, depending on your internet
connection.

```bash
make install
```

#### Running the tests/linters/pre-commit

Linter will be run as par of pre-commit check whenever you are creating new commit.
To run tests use:

```bash
make test
make test-container # Runs inside a docker.io/golang:1.20 container
```

#### Pre-commit

Pre-commit is an important tool for our development workflow, with this tool we
can run a series of pre-defined hooks against our codebase to keep it clean and
maintainable. Here is an example of output from `pre-commit` being run:

```
(.venv3) [rhc-worker-bash]$ pre-commit run --all-files
golangci-lint............................................................Passed
fix end of files.........................................................Passed
trim trailing whitespace.................................................Passed
check for merge conflicts................................................Passed
```

We automatically run `pre-commit` as part of our CI infrastructure as well. If you have a PR it will run and see if everything passes. Sometimes there may be an outage or unexpected result from `pre-commit`, if that happens you can create a new comment on the PR saying:

> pre-commit.ci run

Install `pre-commit` hooks to automatically run when doing `git commit`.

```bash
# installs pre-commit hooks into the repo (included into make install)
pre-commit install --install-hooks
```

Running `pre-commit` against our files

```bash
# run pre-commit hooks for staged files
pre-commit run

# run pre-commit hooks for all files in repo
pre-commit run --all-files

# bypass pre-commit hooks
git commit --no-verify
```

And lastly but not least, if you wish to update our hooks, we can do so by
running the command:

```bash
# bump versions of the pre-commit hooks automatically to the latest available
pre-commit autoupdate
```

If you wish to learn more about all the things that `pre-commit` can do, refer
to their documentation on [how to use
pre-commit](https://pre-commit.com/#usage).

### Writing tests

Tests are an important part of the development process, they guarantee to us
that our code is working in the correct way as expected, and for RHC Worker Bash,
we separate these tests in two categories.

- Unit testing
- Integration testing

TODO

## Additional information

TODO
