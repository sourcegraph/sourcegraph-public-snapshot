TARGETS=xlang-javascript-typescript xlang-php xlang-java

default:
	@echo Please specify a target: $(TARGETS)

all: $(TARGETS)

xlang-javascript-typescript:
	javascript-typescript/build.sh

xlang-php:
	php/build.sh

xlang-java:
	@echo Java build disabled since repo is now private. If you want to build, please manually run
	@echo java/build.sh
