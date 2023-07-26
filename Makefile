.PHONY: \
	install \
	clean \
	build \
	distribution-tarball \
	test \
	test-container \
	publish \
	development \
	coverage \
	coverage-html

# Project constants
VERSION ?= 0.2
PKGNAME ?= rhc-worker-bash
GO_SOURCES := $(wildcard src/*.go)
PYTHON ?= python3
PIP ?= pip3
VENV ?= .venv3
PRE_COMMIT ?= pre-commit
GO_VERSION ?= 1.16

ifdef KEEP_TEST_CONTAINER
	CONTAINER_RM =
else
	CONTAINER_RM = --rm
endif

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


# NOTE: We could also add -race option to add detection for race conditions,
# however that significantly increases time execution
test:
	go test -coverprofile=coverage.out ./...

test-container:
	podman run --replace --name go-test-container $(CONTAINER_RM) -v $(shell pwd):/app:Z -w /app docker.io/golang:$(GO_VERSION) make test

coverage: test
	go tool cover -func=coverage.out

coverage-html: test
	go tool cover -html=coverage.out

publish:
	. $(VENV)/bin/activate; python development/python/mqtt_publish.py

development:
	@podman-compose -f development/podman-compose.yml down
	podman-compose -f development/podman-compose.yml up
