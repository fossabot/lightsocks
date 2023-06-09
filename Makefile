SHELL=/bin/bash
BINARY_NAME := lightsocks
LDFLAGS := "-w -s"

all: windows linux darwin freebsd netbsd openbsd

dev:
	@go mod tidy
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_linux_amd64 cmd/$(BINARY_NAME).go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_windows_amd64.exe cmd/$(BINARY_NAME).go

windows:
	@echo "build windows $(BINARY_NAME)"
	@go mod tidy
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_windows_386.exe cmd/$(BINARY_NAME).go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_windows_amd64.exe cmd/$(BINARY_NAME).go

linux:
	@echo "build linux $(BINARY_NAME)"
	@go mod tidy
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_linux_386 cmd/$(BINARY_NAME).go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_linux_amd64 cmd/$(BINARY_NAME).go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_linux_arm cmd/$(BINARY_NAME).go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_linux_arm64 cmd/$(BINARY_NAME).go
	CGO_ENABLED=0 GOOS=linux GOARCH=ppc64 go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_linux_ppc64 cmd/$(BINARY_NAME).go
	CGO_ENABLED=0 GOOS=linux GOARCH=ppc64le go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_linux_ppc64le cmd/$(BINARY_NAME).go
	CGO_ENABLED=0 GOOS=linux GOARCH=mips go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_linux_mips cmd/$(BINARY_NAME).go
	CGO_ENABLED=0 GOOS=linux GOARCH=mipsle go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_linux_mipsle cmd/$(BINARY_NAME).go
	CGO_ENABLED=0 GOOS=linux GOARCH=mips64 go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_linux_mips64 cmd/$(BINARY_NAME).go
	CGO_ENABLED=0 GOOS=linux GOARCH=mips64le go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_linux_mips64le cmd/$(BINARY_NAME).go

darwin:
	@echo "build darwin $(BINARY_NAME)"
	@go mod tidy
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_darwin_amd64 cmd/$(BINARY_NAME).go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_darwin_arm64 cmd/$(BINARY_NAME).go

freebsd:
	@echo "build freebsd $(BINARY_NAME)"
	@go mod tidy
	CGO_ENABLED=0 GOOS=freebsd GOARCH=386 go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_freebsd_386 cmd/$(BINARY_NAME).go
	CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_freebsd_amd64 cmd/$(BINARY_NAME).go
	CGO_ENABLED=0 GOOS=freebsd GOARCH=arm go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_freebsd_arm cmd/$(BINARY_NAME).go

netbsd:
	@echo "build netbsd $(BINARY_NAME)"
	@go mod tidy
	CGO_ENABLED=0 GOOS=netbsd GOARCH=386 go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_netbsd_386 cmd/$(BINARY_NAME).go
	CGO_ENABLED=0 GOOS=netbsd GOARCH=amd64 go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_netbsd_amd64 cmd/$(BINARY_NAME).go
	CGO_ENABLED=0 GOOS=netbsd GOARCH=arm go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_netbsd_arm cmd/$(BINARY_NAME).go

openbsd:
	@echo "build openbsd $(BINARY_NAME)"
	@go mod tidy
	CGO_ENABLED=0 GOOS=openbsd GOARCH=386 go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_openbsd_386 cmd/$(BINARY_NAME).go
	CGO_ENABLED=0 GOOS=openbsd GOARCH=amd64 go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_openbsd_amd64 cmd/$(BINARY_NAME).go
	CGO_ENABLED=0 GOOS=openbsd GOARCH=arm go build -trimpath -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_openbsd_arm cmd/$(BINARY_NAME).go