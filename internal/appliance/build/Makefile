VERSION=v1.0.$$(( $(shell date +%s) - 1719292685 ))

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
	cat ../helm/values.yaml \
	  | sed -e "s#registry:.*#registry: ${REGISTRY}#" \
	  | sed -e "s/\(.*image:.*\):[^:]*/\1:${VERSION}/" \
	  > ../helm/values.yaml.tmp
	mv ../helm/values.yaml.tmp ../helm/values.yaml
	cat ../helm/Chart.yaml \
	  | sed -e "s/version:.*/version: ${VERSION}/" \
	  | sed -e "s/appVersion:.*/appVersion: \"${VERSION}\"/" \
	  > ../helm/Chart.yaml.tmp
	mv ../helm/Chart.yaml.tmp ../helm/Chart.yaml
