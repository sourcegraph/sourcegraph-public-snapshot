PREFIX=/usr

.PHONY: install node_exporter package package_dep

node_exporter:
	$(MAKE) package PACKAGE_BIN=node_exporter PACKAGE_URL=https://github.com/prometheus/node_exporter/releases/download/0.11.0/node_exporter-0.11.0.linux-amd64.tar.gz VERSION=0.11.0

bin/${PACKAGE_BIN}:
	mkdir -p bin
	cd bin && curl -Ls ${PACKAGE_URL} | tar zx

install: bin/${PACKAGE_BIN}
	install -Dm 755 bin/${PACKAGE_BIN} ${DESTDIR}${PREFIX}/bin/${PACKAGE_BIN}

package: install
	rm -rf pkg
	$(MAKE) install DESTDIR=pkg
	mkdir -p ../dist
	fpm --rpm-os linux \
		-s dir \
		-p ../dist/${PACKAGE_BIN}-${VERSION}.deb \
		-t deb \
		-n ${PACKAGE_BIN} \
		--version ${VERSION} \
		--deb-upstart ${PACKAGE_BIN}.upstart \
		-C pkg \
		.
