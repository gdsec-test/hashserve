bin_name := hashserve
reponame := docker-dcu-local/$(bin_name)
dockerrepo := docker-dcu-local.artifactory.secureserver.net/$(bin_name)
dockerrepohasher := docker-dcu-local.artifactory.secureserver.net/hasher
# https://golang.org/doc/install/source#environment
build_pkg ?= github.com/gdcorp-infosec/hashserve
commit :=
shell := /bin/bash
build_version ?= $(shell git describe --always)
build_date := $(shell date)

# build config
build_goos := darwin linux
build_goarch := amd64
build_branch := origin/master

build_ldflags := -ldflags='-X main.Version=$(build_version)
build_ldflags += -X main.Version=$(build_version)'

build_name := $(bin_name)-$(build_version)
build_targets := $(strip \
	$(foreach goos,$(build_goos), \
		$(foreach goarch,$(build_goarch), \
			$(build_name)-$(goos)-$(goarch)$(if $(findstring windows,$(goos)),.exe,))))

define deploy_k8s
	docker push $(dockerrepo):$(2)
	cd k8s/$(1) && kustomize edit set image $$(docker inspect --format='{{index .RepoDigests 0}}' $(dockerrepo):$(2))
	kubectl --context $(3) apply -k k8s/$(1)
	cd k8s/$(1) && kustomize edit set image $(dockerrepo):$(1)
endef


bail = { printf >&2 "\nError: $(1))\n"; exit 1; }

.PHONY: all
all: check test build

.PHONY: release
release: all
	$(info release...)
	shasum -a 256 build/* > build/release || true
	cat build/release

.PHONY: check
check:
	@command -v go >/dev/null 2>&1 || \
		$(call bail,go command not found - bad build env? GOROOT=$(GOROOT))

.PHONY: unit-test
unit-test:
	go test ./...

.PHONY: testcov
testcov:
	@echo "----- Running tests with coverage -----"
	go test ./... -cover

.PHONY: init
init:
	@echo "----- Make init -----"

build: $(addprefix build/,$(build_targets))
	mkdir -p build
	touch $$PWD

build/$(build_name)%: $(wildcard cmd/$(bin_name)/*.go)
	GOOS=$(strip $(foreach s,$(build_goos),$(findstring $(s),$(@F)))) \
	GOARCH=$(strip $(foreach s,$(build_goarch),$(findstring $(s),$(@F)))) \
		go build -v -o $(@) $(build_ldflags) $(build_pkg)/cmd/$(bin_name)

.PHONY: prep
prep: build
	@echo "----- preparing $(reponame) $(build_version) -----"
	# copy the app code
	cp -p build/$(build_name)-linux-amd64 build/$(bin_name)

.PHONY: dev
dev: prep
	@echo "----- building $(reponame) $(build_version) -----"
	docker build -t $(dockerrepo):dev .

.PHONY: test-build
test-build: prep
	@echo "----- building $(reponame) $(build_version) -----"
	docker build -t $(dockerrepo):test .

.PHONY: prod
prod: prep
	@echo "----- building $(reponame) $(build_version) -----"
	$(eval commit:=$(shell git rev-parse --short HEAD))
	docker build -t $(dockerrepo):$(commit) .

.PHONY: ote
ote: prep
	@echo "----- building $(reponame) $(build_version) -----"
	$(eval commit:=$(shell git rev-parse --short HEAD))
	docker build -t $(dockerrepo):$(commit) .

.PHONY: prod-deploy
prod-deploy: prod
	@echo "----- deploying $(reponame) prod -----"
	$(eval commit:=$(shell git rev-parse --short HEAD))
	read -p "About to deploy a production image. Are you sure? (Y/N): " response ; \
	if [[ $$response == 'N' || $$response == 'n' ]] ; then exit 1 ; fi
	if [[ `git status --porcelain | wc -l` -gt 0 ]] ; then echo "You must stash your changes before proceeding" ; exit 1 ; fi
	$(call deploy_k8s,prod,$(commit),prod-cset)

.PHONY: dev-deploy
dev-deploy: dev
	@echo "----- deploying $(reponame) dev -----"
	$(call deploy_k8s,dev,dev,dev-cset)

.PHONY: test-deploy
test-deploy: test-build
	@echo "----- deploying $(reponame) test -----"
	$(call deploy_k8s,test,test,test-cset)

.PHONY: ote-deploy
ote-deploy: ote
	@echo "----- deploying $(reponame) ote -----"
	$(eval commit:=$(shell git rev-parse --short HEAD))
	if [[ `git status --porcelain | wc -l` -gt 0 ]] ; then echo "You must stash your changes before proceeding" ; exit 1 ; fi
	$(call deploy_k8s,ote,$(commit),ote-cset)

PHONY: clean
clean:
	@echo "----- Cleaning $(reponame) project -----"
	rm -f $(wildcard build/$(bin_name)*)