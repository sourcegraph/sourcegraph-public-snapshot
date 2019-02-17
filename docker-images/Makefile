images += alpine

default:
	@echo Possible targets: $(images)

$(images):
	$(eval tag := $(shell ./version.sh $@))
	$(eval image := sourcegraph/$@)
	docker build -t $(image):$(tag) $@
	docker tag $(image):$(tag) $(image):latest
	docker push $(image):$(tag)
	docker push $(image):latest

.PHONY: $(images)
