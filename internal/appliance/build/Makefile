VERSION=1.0.$$(( $(shell date +%s) - 1719292685 ))

REGISTRY=us-central1-docker.pkg.dev/nelsona-sandbox/nelsona-dev-images

.PHONY: all

all:
	make .version VERSION=${VERSION} REGISTRY=${REGISTRY}
	make build VERSION=${VERSION} REGISTRY=${REGISTRY}
	make pkg VERSION=${VERSION} REGISTRY=${REGISTRY}

build:
	@echo "---------------------------------------"
	@echo "Building ${VERSION}..."
	@echo "---------------------------------------"
	make build.selfupdater VERSION=${VERSION} REGISTRY=${REGISTRY}
	@echo "---------------------------------------"
	@echo "Build ${VERSION} complete"
	@echo "---------------------------------------"

pkg:
	@echo "---------------------------------------"
	@echo "Packging ${VERSION}..."
	@echo "---------------------------------------"
	make pkg.selfupdater VERSION=${VERSION} REGISTRY=${REGISTRY}
	make helm VERSION=${VERSION} REGISTRY=${REGISTRY}
	@echo "---------------------------------------"
	@echo "Package ${VERSION} complete"
	@echo "---------------------------------------"

.version:
	@echo "======================================="
	@echo "Building version ${VERSION}"
	@echo "======================================="

build.selfupdater:
	cd ../selfupdate && go mod download
	cd ../selfupdate && GOOS=linux GOARCH=amd64 go build -o ../dist/selfupdater ./cmd

pkg.selfupdater:
	make -C selfupdate VERSION=${VERSION} REGISTRY=${REGISTRY}

helm:
	@echo "Updating Helm to version ${VERSION}"
	cat ../helm/selfupdater/values.yaml \
	  | sed -e "s#registry:.*#registry: ${REGISTRY}#" \
	  | sed -e "s/\(.*image:.*\):[^:]*/\1:${VERSION}/" \
	  > ../helm/selfupdater/values.yaml.tmp
	mv ../helm/selfupdater/values.yaml.tmp ../helm/selfupdater/values.yaml
	cat ../helm/selfupdater/Chart.yaml \
	  | sed -e "s/version:.*/version: ${VERSION}/" \
	  | sed -e "s/appVersion:.*/appVersion: ${VERSION}/" \
	  > ../helm/selfupdater/Chart.yaml.tmp
	mv ../helm/selfupdater/Chart.yaml.tmp ../helm/selfupdater/Chart.yaml
