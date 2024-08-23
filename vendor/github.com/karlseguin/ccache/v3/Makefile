.PHONY: t
t:
	go test -race -count=1 ./...

.PHONY: f
f:
	go fmt ./...


.PHONY: c
c:
	go test -race -covermode=atomic ./... -coverprofile=cover.out && \
# 	go tool cover -html=cover.out && \
	go tool cover -func cover.out \
		| grep -vP '[89]\d\.\d%' | grep -v '100.0%' \
		|| true

	rm cover.out
