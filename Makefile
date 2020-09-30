
IGNORED:=$(shell bash -c "source .metadata.sh ; env | sed 's/=/:=/;s/^/export /' > .metadata.make")

ifeq ($(VERSION),)
	include .metadata.make
else
	# Preserve the passed-in version & iteration (homebrew).
	_VERSION:=$(VERSION)
	_ITERATION:=$(ITERATION)
	include .metadata.make
	VERSION:=$(_VERSION)
	ITERATION:=$(_ITERATION)
endif

all: hciscan

build: $(BINARY)
$(BINARY): main.go
	GOOS=$(OS) GOARCH=$(ARCH) go build -o $(BINARY) -ldflags "-w -s $(VERSION_LDFLAGS)"

exe: $(BINARY).amd64.exe
windows: $(BINARY).amd64.exe
$(BINARY).amd64.exe: main.go
	# Building windows 64-bit x86 binary.
	GOOS=windows GOARCH=amd64 go build -o $@ -ldflags "-w -s $(VERSION_LDFLAGS)"

docker:
	docker buildx create --name hciscan
	docker buildx use hciscan
	docker buildx inspect --bootstrap
	docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7,linux/386 -t mabunixda/hciscan --push .

