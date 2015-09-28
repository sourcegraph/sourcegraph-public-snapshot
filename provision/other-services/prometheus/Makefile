ROOT=$(shell git rev-parse --show-toplevel)
SSH=ssh -i $(HOME)/.ssh/sg-dev.pem
RSYNC=rsync -Ccrvz -e "$(SSH)"
TARGET=ubuntu@metrics.sgdev.org

.PHONY: deploy clone-private

deploy: clone-private env
	$(RSYNC) --exclude 'data' . $(TARGET):prometheus
	$(SSH) $(TARGET) 'cd prometheus && source ./env && docker-compose up -d --force-recreate'

env: $(ROOT)/conf/private/prometheus
	cp $(ROOT)/conf/private/prometheus env

clone-private:
	cd $(ROOT) && $(MAKE) clone-private
