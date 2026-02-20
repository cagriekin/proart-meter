PREFIX ?= /usr
BINDIR = $(PREFIX)/bin
CONFDIR = /etc/proart-meter
SYSTEMDDIR = /usr/lib/systemd/system

.PHONY: build clean install uninstall

build:
	go build -o proart-meter ./cmd/proart-meter

clean:
	rm -f proart-meter

install: build
	install -Dm755 proart-meter $(DESTDIR)$(BINDIR)/proart-meter
	install -Dm644 config.yaml $(DESTDIR)$(CONFDIR)/config.yaml
	install -Dm644 proart-meter.service $(DESTDIR)$(SYSTEMDDIR)/proart-meter.service

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/proart-meter
	rm -f $(DESTDIR)$(SYSTEMDDIR)/proart-meter.service
