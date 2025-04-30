PKGNAME := foreman_ygg_worker
LIBEXECDIR := /usr/libexec
WORKER_GROUP := yggdrasil-worker

ifeq ($(origin VERSION), undefined)
	VERSION := 0.3.1
endif

go_sources := $(wildcard src/*.go)

build: $(go_sources)
	mkdir -p build
	CGO_ENABLED=0 go build -o build/$(PKGNAME) $^

.PHONY: data
data: build/data/com.redhat.Yggdrasil1.Worker1.foreman.conf build/data/com.redhat.Yggdrasil1.Worker1.foreman.service

.PHONY: install
install: build data
	install -D -m 755 build/$(PKGNAME) $(DESTDIR)$(LIBEXECDIR)/$(PKGNAME)
	install -D -m 644 build/data/com.redhat.Yggdrasil1.Worker1.foreman.conf $(DESTDIR)/usr/share/dbus-1/system.d/com.redhat.Yggdrasil1.Worker1.foreman.conf
	install -D -m 644 data/dbus_com.redhat.Yggdrasil1.Worker1.foreman.service $(DESTDIR)/usr/share/dbus-1/system-services/com.redhat.Yggdrasil1.Worker1.foreman.service
	install -D -m 644 build/data/com.redhat.Yggdrasil1.Worker1.foreman.service $(DESTDIR)/usr/lib/systemd/system/com.redhat.Yggdrasil1.Worker1.foreman.service

clean:
	rm -rf build

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
		--exclude=.packit.yml \
		--transform s/^\./$(PKGNAME)-$(VERSION)/ \
		. && mv /tmp/$(PKGNAME)-$(VERSION).tar.gz .
	rm -rf ./vendor
	@echo $(PKGNAME)-$(VERSION).tar.gz

test:
	go test src/*

vet:
	go vet src/*

build/data/%: data/%.in
	mkdir -p $(@D)
	sed \
		-e 's,[@]libexecdir[@],$(LIBEXECDIR),g' \
		-e 's,[@]worker_group[@],$(WORKER_GROUP),g' \
		-e 's,[@]executable[@],$(PKGNAME),g' \
		$< > $@
