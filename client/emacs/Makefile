# https://github.com/rolandwalker/emacs-travis

EMACS=emacs

EMACS_CLEAN=-Q
EMACS_BATCH=$(EMACS_CLEAN) --batch
TESTS=

CURL=curl --silent
WORK_DIR=$(shell pwd)
PACKAGE_NAME=sourcegraph
AUTOLOADS_FILE=$(PACKAGE_NAME)-autoloads.el

.PHONY: build autoloads test-autoloads test clean install-circleci

build:
	$(EMACS) $(EMACS_BATCH) --eval             \
	    "(progn                                \
	      (setq byte-compile-error-on-warn t)  \
	      (batch-byte-compile))" *.el

autoloads:
	$(EMACS) $(EMACS_BATCH) --eval                       \
	    "(progn                                          \
	      (setq generated-autoload-file \"$(WORK_DIR)/$(AUTOLOADS_FILE)\") \
	      (update-directory-autoloads \"$(WORK_DIR)\"))"

test-autoloads: autoloads
	@$(EMACS) $(EMACS_BATCH) -L . -l "./$(AUTOLOADS_FILE)"      || \
	 ( echo "failed to load autoloads: $(AUTOLOADS_FILE)" && false )

test: build test-autoloads

clean:
	@rm -f $(AUTOLOADS_FILE) *.elc *~ */*.elc */*~

# EMACS=emacs24 or EMACS=emacs-snapshot
install-circleci:
	@if [ "$EMACS" = 'emacs24' ]; then \
		sudo add-apt-repository -y ppa:cassou/emacs && \
		sudo apt-get -qq update && \
		sudo apt-get -qq -f install && \
		sudo apt-get -qq install emacs24 emacs24-el; \
		fi
	@if [ "$EMACS" = 'emacs-snapshot' ]; then \
		sudo add-apt-repository -y ppa:ubuntu-elisp/ppa &&
		sudo apt-get -qq update && \
		sudo apt-get -qq -f install && \
		sudo apt-get -qq install emacs-snapshot && \
		sudo apt-get -qq install emacs-snapshot-el; \
		fi