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

# Build constants
VERSION ?= 0.8
PKGNAME ?= rhc-worker-script
GO_SOURCES := $(wildcard src/*.go)
GO_VERSION ?= 1.21

BUILDFLAGS ?=
LDFLAGS ?=
ifeq ($(shell find . -name vendor), ./vendor)
BUILDFLAGS += -mod=vendor
endif

# -----------------------------------------------------------------------------

# Development constants
_PYTHON ?= python3
_PIP ?= pip3
_VENV ?= .venv3
_PRE_COMMIT ?= pre-commit
_CLIENT_ID ?= 00000000-0000-0000-0000-0000000000000
_SERVED_FILENAME ?= example_bash.yml

ifdef KEEP_TEST_CONTAINER
	_CONTAINER_RM =
else
	_CONTAINER_RM = --rm
endif

# -----------------------------------------------------------------------------

all: clean build

install: .install .pre-commit

.install:
	virtualenv --system-site-packages --python $(_PYTHON) $(_VENV); \
	. $(_VENV)/bin/activate; \
	$(_PIP) install --upgrade -r ./development/python/requirements.txt; \
	touch $@

.pre-commit:
	$(_PRE_COMMIT) install --install-hooks
	touch $@

clean:
	rm -rf build

build: $(GO_SOURCES)
	mkdir -p build
	go build $(BUILDFLAGS) -ldflags "$(LDFLAGS)" -o build/rhc-script-worker $^

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
		--exclude=development \
		--transform s/^\./$(PKGNAME)-$(VERSION)/ \
		. && mv /tmp/$(PKGNAME)-$(VERSION).tar.gz .
	rm -rf ./vendor


# NOTE: We could also add -race option to add detection for race conditions,
# however that significantly increases time execution
test:
	go test -coverprofile=coverage.out ./...

test-container:
	podman run --replace --name go-test-container $(_CONTAINER_RM) -v $(shell pwd):/app:Z -w /app docker.io/golang:$(GO_VERSION) make test

coverage: test
	go tool cover -func=coverage.out

coverage-html: test
	go tool cover -html=coverage.out

publish:
	. $(_VENV)/bin/activate; python development/python/mqtt_publish.py $(_CLIENT_ID) $(_SERVED_FILENAME)

development:
	@podman-compose -f development/podman-compose.yml down
	podman-compose -f development/podman-compose.yml up -d
