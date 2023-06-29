.PHONY: \
	install \
	clean \
	build \
	distribution-tarball \
	test \
	publish \
	development

# Project constants
VERSION ?= 0.1
PKGNAME ?= rhc-worker-bash
GO_SOURCES := $(wildcard src/*.go)
PYTHON ?= python3
PIP ?= pip3
VENV ?= .venv3
PRE_COMMIT ?= pre-commit

all: clean build

install: .install .pre-commit

.install:
	virtualenv --system-site-packages --python $(PYTHON) $(VENV); \
	. $(VENV)/bin/activate; \
	$(PIP) install --upgrade -r ./development/python/requirements.txt; \
	touch $@

.pre-commit:
	$(PRE_COMMIT) install --install-hooks
	touch $@

clean:
	rm -rf build

build: $(GO_SOURCES)
	mkdir -p build
	CGO_ENABLED=0 go build -o build/rhc-bash-worker $^

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
	. $(VENV)/bin/activate; python development/python/mqtt_publish.py

development:
	@podman-compose -f development/podman-compose.yml down
	podman-compose -f development/podman-compose.yml up
