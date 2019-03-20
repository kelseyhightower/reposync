.PHONY: build

build:
	GOOS=linux go build -o function .

package: build
	$(eval TAG := $(shell git describe --abbrev=0 --tags))
	mkdir reposync-cloud-function-$(TAG)
	mv function reposync-cloud-function-$(TAG)/
	cp index.js reposync-cloud-function-$(TAG)/
	zip -r -9 reposync-cloud-function-$(TAG).zip reposync-cloud-function-$(TAG)/
	rm -fr reposync-cloud-function-$(TAG)/