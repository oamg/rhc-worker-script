.PHONY: \
	clean \
	build \
	distribution-tarball \
	test \
	test-container \
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

ifdef KEEP_TEST_CONTAINER
	_CONTAINER_RM =
else
	_CONTAINER_RM = --rm
endif

# -----------------------------------------------------------------------------

all: clean .pre-commit build

.pre-commit:
	pre-commit install --install-hooks
	touch $@

clean:
	@rm -rf build/
	@find . -name '.pre-commit' -exec rm -fr {} +
	@find . -name 'coverage.out' -exec rm -rf {} +


build: $(GO_SOURCES)
	mkdir -p build
	CGO_ENABLED=0 go build $(BUILDFLAGS) -ldflags "$(LDFLAGS)" -o build/rhc-script-worker $^

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
