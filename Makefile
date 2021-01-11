bin_name := hashserve
reponame := docker-dcu-local/$(bin_name)
buildroot := $(HOME)/dockerbuild/$(reponame)
dockerrepo := docker-dcu-local.artifactory.secureserver.net/$(bin_name)
# https://golang.org/doc/install/source#environment
build_pkg ?= github.secureserver.net/digital-crimes/hashserve

build_version ?= $(shell git describe --always)
build_date := $(shell date)

# build config
build_goos := darwin linux windows
build_goarch := amd64

build_ldflags := -ldflags='-X main.Version=$(build_version)
build_ldflags += -X main.Version=$(build_version)'

build_name := $(bin_name)-$(build_version)
build_targets := $(strip \
	$(foreach goos,$(build_goos), \
		$(foreach goarch,$(build_goarch), \
			$(build_name)-$(goos)-$(goarch)$(if $(findstring windows,$(goos)),.exe,))))

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

.PHONY: test
test:
	go test ./...

build: $(addprefix build/,$(build_targets))
	mkdir -p build
	touch $$PWD

build/$(build_name)%: $(wildcard cmd/$(bin_name)/*.go)
	GOOS=$(strip $(foreach s,$(build_goos),$(findstring $(s),$(@F)))) \
	GOARCH=$(strip $(foreach s,$(build_goarch),$(findstring $(s),$(@F)))) \
		go build -i -v -o $(@) $(build_ldflags) $(build_pkg)/cmd/$(bin_name)

.PHONY: prep
prep: build
	@echo "----- preparing $(reponame) $(build_version) -----"
	mkdir -p $(buildroot) && rm -rf $(buildroot)/*
	# copy the app code to the build root
	cp -p build/$(build_name)-linux-amd64 $(buildroot)/$(bin_name)
	cp -rp k8s $(buildroot)
	cp Dockerfile $(buildroot)

.PHONY: dev
dev: prep
	@echo "----- building $(reponame) $(build_version) -----"
	sed -ie 's/THIS_STRING_IS_REPLACED_DURING_BUILD/$(build_date)/g' $(buildroot)/k8s/dev/deployment.yaml
	docker build -t $(dockerrepo):dev $(buildroot)

.PHONY: prod
prod: prep
	@echo "----- building $(reponame) $(build_version) -----"
	sed -ie 's/THIS_STRING_IS_REPLACED_DURING_BUILD/$(build_date)/g' $(buildroot)/k8s/prod/deployment.yaml
	docker build -t $(dockerrepo):prod $(buildroot)

.PHONY: prod-deploy
prod-deploy: prod
	@echo "----- deploying $(reponame) prod -----"
	docker push $(dockerrepo):prod
	kubectl --context prod-dcu apply -f $(buildroot)/k8s/prod/deployment.yaml --record

.PHONY: dev-deploy
dev-deploy: dev
	@echo "----- deploying $(reponame) dev -----"
	docker push $(dockerrepo):dev
	kubectl --context dev-dcu apply -f $(buildroot)/k8s/dev/deployment.yaml --record

PHONY: clean
clean:
	@echo "----- Cleaning $(reponame) project -----"
	rm -f $(wildcard build/$(bin_name)*)