
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

all: docker

build: $(BINARY)
$(BINARY): main.go
	GOOS=$(OS) GOARCH=$(ARCH) go build -o $(BINARY) -ldflags "-w -s $(VERSION_LDFLAGS)"

exe: $(BINARY).amd64.exe
windows: $(BINARY).amd64.exe
$(BINARY).amd64.exe: main.go
	# Building windows 64-bit x86 binary.
	GOOS=windows GOARCH=amd64 go build -o $@ -ldflags "-w -s $(VERSION_LDFLAGS)"


docker:
#	docker buildx create --name hciscan
	docker buildx use hciscan
	docker buildx inspect --bootstrap
	docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7,linux/386 -t mabunixda/hciscan --push .

docker.linux.amd64:
	docker build -f Dockerfile \
		--build-arg "BUILD_DATE=$(DATE)" \
		--build-arg "COMMIT=$(COMMIT)" \
		--build-arg "VERSION=$(VERSION)-$(ITERATION)" \
		--build-arg "LICENSE=$(LICENSE)" \
		--build-arg "DESC=$(DESC)" \
		--build-arg "URL=$(URL)" \
		--build-arg "VENDOR=$(VENDOR)" \
		--build-arg "AUTHOR=$(MAINT)" \
		--build-arg "BINARY=$(BINARY)" \
		--build-arg "SOURCE_URL=$(SOURCE_URL)" \
		--tag "mabunixda/$(BINARY):${VERSION}-amd64" .

docker.linux.arm64:
	docker build -f Dockerfile.arm64 \
		--build-arg "BUILD_DATE=$(DATE)" \
		--build-arg "COMMIT=$(COMMIT)" \
		--build-arg "VERSION=$(VERSION)-$(ITERATION)" \
		--build-arg "LICENSE=$(LICENSE)" \
		--build-arg "DESC=$(DESC)" \
		--build-arg "URL=$(URL)" \
		--build-arg "VENDOR=$(VENDOR)" \
		--build-arg "AUTHOR=$(MAINT)" \
		--build-arg "BINARY=$(BINARY)" \
		--build-arg "SOURCE_URL=$(SOURCE_URL)" \
		--build-arg "ARCH=arm64" \
		--tag "mabunixda/$(BINARY):${VERSION}-arm64" .
