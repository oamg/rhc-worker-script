.PHONY: \
	install
	clean
	build
	distribution-tarball
	test
	publish
	broker

# Project constants
VERSION ?= 0.1
PKGNAME ?= rhc-bash-worker
GO_SOURCES := $(wildcard src/*.go)
PYTHON ?= python3
PIP ?= pip3
VENV ?= .venv3
PRE_COMMIT ?= pre-commit

# Let the user specify PODMAN at the CLI, otherwise try to autodetect a working podman
ifndef PODMAN
	PODMAN := $(shell podman run --rm alpine echo podman 2> /dev/null)
	ifndef PODMAN
		DUMMY := $(warning podman is not detected. Majority of commands will not work. Please install and verify that podman --version works.)
	endif
endif

all: clean build

install: .install .pre-commit

.install:
	virtualenv --system-site-packages --python $(PYTHON) $(VENV); \
	. $(VENV)/bin/activate; \
	$(PIP) install --upgrade -r ./scripts/requirements.txt; \
	touch $@

.pre-commit:
	$(PRE_COMMIT) install --install-hooks
	touch $@

clean:
	@rm -rf build

build: $(GO_SOURCES)
	mkdir -p build
	CGO_ENABLED=0 go build -o build/$(PKGNAME) $^

distribution-tarball:
	go mod vendor
	tar --create \
		--gzip \
		--file /tmp/$(PKGNAME)-$(VERSION).tar.gz \
		--exclude=.git \
		--exclude=.vscode \
		--exclude=.github \
		--exclude=.gitignore \
		--exclude=.copr \
		--transform s/^\./$(PKGNAME)-$(VERSION)/ \
		. && mv /tmp/$(PKGNAME)-$(VERSION).tar.gz .
	rm -rf ./vendor

test:
	go test src/*

publish:
	. $(VENV)/bin/activate; python python/mqtt_publish.py

broker:
	@$(PODMAN) run -d -it -p 1883:1883 -p 9001:9001 -v $(PWD)/mosquitto/mosquitto.conf:/mosquitto/config/mosquitto.conf:Z eclipse-mosquitto
