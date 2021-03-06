.PHONY: all docker install

DOCKER=docker
GO=go
INSTALL=install
SYSTEMCTL=systemctl

all: HKUST-MATH-Seminar-to-iCal x86_64-linux-gnu-HKUST-MATH-Seminar-to-iCal

docker: x86_64-linux-gnu-HKUST-MATH-Seminar-to-iCal
	$(DOCKER) build -t seminar .

install: all docker
	$(INSTALL) -Dm0644 docker-seminar.service /etc/systemd/system/docker-seminar.service
	$(SYSTEMCTL) daemon-reload || true

HKUST-MATH-Seminar-to-iCal: ical.go main.go pattern.go util.go
	$(GO) build -o $@

x86_64-linux-gnu-HKUST-MATH-Seminar-to-iCal: ical.go main.go pattern.go util.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -o $@
