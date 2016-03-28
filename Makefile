CGO_ENABLED=0

RELAYD      := relayd
VERSION     := 0.0.1

SOURCES			:= *.go */*.go _vendor
OS				:= $(shell /usr/bin/lsb_release -si 2> /dev/null)

space :=
space +=
comma := ,

GO15VENDOREXPERIMENT=0
GOM_VENDOR_NAME=_vendor
export GO15VENDOREXPERIMENT
export GOM_VENDOR_NAME

# symlinks confuse go tools, let's not mess with it and use -L
all: fmt lint $(RELAYD) test

.PHONY: clean
clean:
	@echo Cleaning $(RELAYD)...
	@rm -rf build relayd*.deb relayd*.rpm relayd _vendor

deps:
	@echo Getting dependencies...
	@go get github.com/mattn/gom
	@gom install > /dev/null

$(RELAYD): $(SOURCES) deps
	@echo Building $(RELAYD)...
	@gom build -a -tags netgo -ldflags '-w' -o $(RELAYD) .

test: tests
tests: deps
	@echo Testing $(RELAYD)
	@gom test -cover ./...

fmt: deps $(SOURCES)
	@gom fmt ./...

vet: deps $(SOURCES)
	@echo Vetting $(RELAYD) sources...
	@gom vet ./...

lint: deps $(SOURCES)
	@echo Linting $(RELAYD) sources...
	@_vendor/bin/golint ./...

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
