RELAYD      := relayd
VERSION     := 0.0.1
SRCDIR         := src
PKGS        := \
	$(RELAYD) \
	$(RELAYD)/listener \
	$(RELAYD)/forwarder \
	$(RELAYD)/internalserver \

SOURCES        := $(foreach pkg, $(PKGS), $(wildcard $(SRCDIR)/$(pkg)/*.go))
OS	       := $(shell /usr/bin/lsb_release -si 2> /dev/null)

space :=
space +=
comma := ,

# symlinks confuse go tools, let's not mess with it and use -L
GOPATH  := $(shell pwd -L)
GOBIN   := $(GOPATH)/bin
export GOPATH

PATH := $(GOBIN):$(PATH)
export PATH

all: clean fmt lint $(RELAYD) test

.PHONY: clean
clean:
	@echo Cleaning $(RELAYD)...
	@rm -f $(RELAYD) bin/$(RELAYD)
	@rm -rf pkg/*/$(RELAYD)
	@rm -rf build relayd*.deb relayd*.rpm

deps:
	@echo Getting dependencies...
	@go get github.com/mattn/gom
	@bin/gom install > /dev/null

$(RELAYD): $(SOURCES) deps
	@echo Building $(RELAYD)...
	@bin/gom build -o bin/$(RELAYD) $@

test: tests
tests: deps
	@echo Testing $(RELAYD)
	@for pkg in $(PKGS); do \
		bin/gom test -cover $$pkg || exit 1;\
	done

coverage_report: deps
	@echo Creating a coverage rport for $(RELAYD)
	@$(foreach pkg, $(PKGS), bin/gom test -coverprofile=coverage.out -coverpkg=$(subst $(space),$(comma),$(PKGS)) $(pkg);)
	@gom tool cover -html=coverage.out


fmt: deps $(SOURCES)
	@$(foreach pkg, $(PKGS), bin/gom fmt $(pkg);)

vet: deps $(SOURCES)
	@echo Vetting $(RELAYD) sources...
	@$(foreach pkg, $(PKGS), bin/gom vet $(pkg);)

lint: deps $(SOURCES)
	@echo Linting $(RELAYD) sources...
	@$(foreach src, $(SOURCES), _vendor/bin/golint $(src);)

cyclo: deps $(SOURCES)
	@echo Checking code complexity...
	@_vendor/bin/gocyclo $(SOURCES)

pkg: package
package: clean $(RELAYD)
	@echo Packaging for $(OS)
	@mkdir -p build/usr/bin build/usr/share/relayd build/etc
	@cp bin/relayd build/usr/bin/
	@cp deb/bin/run-* build/usr/bin/
ifeq ($(OS),Ubuntu)
	@fpm -s dir \
		-t deb \
		--name $(RELAYD) \
		--version $(VERSION) \
		--description "message forwarder" \
		--depends python \
		--deb-user "relayd" \
		--deb-group "relayd" \
		--deb-default "deb/etc/relayd" \
		--deb-upstart "deb/etc/init/relayd" \
		--before-install "deb/before_install.sh" \
		--before-remove "deb/before_rm.sh" \
		--after-remove "deb/post_rm.sh" \
		-C build .
# CentOS 7 Only
else ifeq ($(OS),CentOS)
	@fpm -s dir \
		-t rpm \
		--name $(RELAYD) \
		--version $(VERSION) \
		--description "message forwarder" \
		--depends python \
		--rpm-user "relayd" \
		--rpm-group "relayd" \
                --before-install "rpm/before_install.sh" \
		--before-remove "rpm/before_rm.sh" \
		-C build . \
		../rpm/relayd.systemd=/etc/systemd/system/relayd.service \
                ../rpm/relayd.sysconfig=/etc/sysconfig/relayd
else
	@echo "OS not supported"
endif
