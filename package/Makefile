VERSION=snapshot
LINUX_BIN=../release/${VERSION}/linux-amd64
DIST_DIR:=$(shell pwd)/dist

.PHONY: package clean test test_centos test_ubuntu package_dep 

package: linux/bin/src package_dep
	docker run -t --rm=true -v ${DIST_DIR}:/release/dist sourcegraph/pkg VERSION=${VERSION}

package_dep:
	docker build -t sourcegraph/pkg -f linux/Dockerfile linux
	@mkdir -p dist

clean:
	rm -fv dist/*${VERSION}.* linux/bin/src

test: test_centos test_ubuntu

test_centos: dist/src-${VERSION}.rpm
	docker build -t sourcegraph/centostest -f linux/Dockerfile.centos.test linux
	docker run -t --rm=true -v ${DIST_DIR}:/workdir/dist sourcegraph/centostest rpm -ivh dist/src-${VERSION}.rpm

test_ubuntu: dist/src-${VERSION}.deb
	docker build -t sourcegraph/ubuntutest -f linux/Dockerfile.ubuntu.test linux
	docker run -t --rm=true -v ${DIST_DIR}:/workdir/dist sourcegraph/ubuntutest dpkg -i dist/src-${VERSION}.deb

linux/bin/src: ${LINUX_BIN}
	@mkdir -p linux/bin
	cp ${LINUX_BIN} linux/bin/src


${LINUX_BIN}:
ifdef CREATE_RELEASE
	@echo ${LINUX_BIN} is missing, so creating a release from the current checkout
	cd release && $(MAKE) PACKAGEFLAGS='--os linux ${VERSION}'
else
	@echo "Please download src-${VERSION} to ${LINUX_BIN}.\nOr we can automatically create a release from the current checkout with\n\n  make VERSION=${VERSION} CREATE_RELEASE=t\n"
	@exit 1
endif
