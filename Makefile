TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=doppler.com
NAMESPACE=core
NAME=doppler
BINARY=terraform-provider-${NAME}
VERSION=dev
OS_ARCH=darwin_amd64

default: install

build:
	go build \
		-ldflags="-X terraform-provider-doppler/doppler.ProviderVersion=dev-$(shell git rev-parse --abbrev-ref HEAD)-$(shell git rev-parse --short HEAD)" \
		-o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test: 
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc: 
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

# https://github.com/hashicorp/terraform-plugin-docs
tfdocs:
	tfplugindocs
