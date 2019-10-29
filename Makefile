.PHONY: all docker install

all: HKUST-MATH-Seminar-to-iCal

docker: x86_64-linux-gnu-HKUST-MATH-Seminar-to-iCal
	docker build -t seminar .

install: all docker
	install -Dm0644 docker-seminar.service /etc/systemd/system/docker-seminar.service

HKUST-MATH-Seminar-to-iCal: *.go
	go build -o $@

x86_64-linux-gnu-HKUST-MATH-Seminar-to-iCal: *.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $@
