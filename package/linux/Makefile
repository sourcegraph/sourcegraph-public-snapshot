PREFIX=/usr
VERSION=snapshot
INIT_FLAVOUR=sysvinit
DIST=dist
# Debian needs a version which starts with a digit
METADATA_VERSION=$(subst snapshot,0.0-snapshot,${VERSION})

.PHONY: all install install-sysvinit install-upstart package

all: ${DIST}/src-${VERSION}.rpm ${DIST}/src-${VERSION}.deb

install: install-${INIT_FLAVOUR}
	install -Dm 755 bin/src ${DESTDIR}${PREFIX}/bin/src
	install -Dm 644 default.ini ${DESTDIR}/etc/sourcegraph/config.ini
	install -Dm 644 default.env ${DESTDIR}/etc/sourcegraph/config.env

install-sysvinit:
	@#We vendor in daemonize since most distros do not ship it
	install -Dm 755 bin/daemonize ${DESTDIR}${PREFIX}/bin/daemonize.sourcegraph
	install -Dm 755 init/sysvinit-src ${DESTDIR}/etc/init.d/src

install-upstart:
	@#We do nothing since we pass in --deb-upstart to fpm

${DIST}/src-${VERSION}.rpm:
	$(MAKE) package PACKAGE_EXT=rpm INIT_FLAVOUR=sysvinit

${DIST}/src-${VERSION}.deb:
	$(MAKE) package PACKAGE_EXT=deb INIT_FLAVOUR=upstart

package:
	rm -rf pkg
	$(MAKE) install DESTDIR=pkg
	mkdir -p ${DIST}/${METADATA_VERSION}
	fpm --rpm-os linux \
		-s dir \
		-p ${DIST}/${METADATA_VERSION}/src.${PACKAGE_EXT} \
		-t ${PACKAGE_EXT} \
		-n src \
		--version ${METADATA_VERSION} \
		--config-files /etc \
		--before-install before_install.sh \
		--after-install after_install.sh \
		--deb-upstart init/src.upstart \
		--depends git \
		-C pkg \
		.

	# Production requires a separate upstart script to rely on Docker starting.
	#
	# TODO(slimsag): this can be removed after we are on GCE.
	cp init/src-docker.upstart /tmp/src.upstart
	fpm --rpm-os linux \
		-s dir \
		-p ${DIST}/${METADATA_VERSION}/src-docker.${PACKAGE_EXT} \
		-t ${PACKAGE_EXT} \
		-n src \
		--version ${METADATA_VERSION} \
		--config-files /etc \
		--before-install before_install.sh \
		--after-install after_install.sh \
		--deb-upstart /tmp/src.upstart \
		--depends git \
		-C pkg \
		.
