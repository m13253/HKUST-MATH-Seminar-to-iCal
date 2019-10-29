.PHONY: docker

all: HKUST-MATH-Seminar-to-iCal docker

HKUST-MATH-Seminar-to-iCal: *.go
	go build -o $@

x86_64-linux-gnu-HKUST-MATH-Seminar-to-iCal: *.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $@

docker: x86_64-linux-gnu-HKUST-MATH-Seminar-to-iCal
	docker build -t seminar .
