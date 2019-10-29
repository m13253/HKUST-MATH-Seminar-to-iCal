.PHONY: all docker install

DOCKER=docker
INSTALL=install
GO=go

all: HKUST-MATH-Seminar-to-iCal

docker: x86_64-linux-gnu-HKUST-MATH-Seminar-to-iCal
	$(DOCKER) build -t seminar .

install: all docker
	$(INSTALL) -Dm0644 docker-seminar.service /etc/systemd/system/docker-seminar.service

HKUST-MATH-Seminar-to-iCal: *.go
	$(GO) build -o $@

x86_64-linux-gnu-HKUST-MATH-Seminar-to-iCal: *.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -o $@
