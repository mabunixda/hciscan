
all: systemd

build: 
	go build -o hciscan


.PHONY: install
install: build
	cp hciscan /usr/local/sbin
	chown root:root /usr/local/sbin/hciscan

.PHONY: systemd
systemd: install
	cp hciscan.service /etc/systemd/system
	systemctl enable hciscan.service
