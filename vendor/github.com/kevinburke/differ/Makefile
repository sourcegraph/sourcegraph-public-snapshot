STATICCHECK := $(GOPATH)/bin/staticcheck
RELEASE := $(GOPATH)/bin/github-release

$(STATICCHECK):
	go get -u honnef.co/go/tools/cmd/staticcheck

vet: $(STATICCHECK)
	go vet ./...
	$(STATICCHECK) ./...

test: vet
	go test ./...

$(RELEASE): test
	go get -u github.com/aktau/github-release

release: $(RELEASE)
ifndef version
	@echo "Please provide a version"
	exit 1
endif
ifndef GITHUB_TOKEN
	@echo "Please set GITHUB_TOKEN in the environment"
	exit 1
endif
	git tag $(version)
	git push origin --tags
	mkdir -p releases/$(version)
	GOOS=linux GOARCH=amd64 go build -o releases/$(version)/differ-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o releases/$(version)/differ-darwin-amd64 .
	GOOS=windows GOARCH=amd64 go build -o releases/$(version)/differ-windows-amd64 .
	# these commands are not idempotent so ignore failures if an upload repeats
	$(RELEASE) release --user kevinburke --repo differ --tag $(version) || true
	$(RELEASE) upload --user kevinburke --repo differ --tag $(version) --name differ-linux-amd64 --file releases/$(version)/differ-linux-amd64 || true
	$(RELEASE) upload --user kevinburke --repo differ --tag $(version) --name differ-darwin-amd64 --file releases/$(version)/differ-darwin-amd64 || true
	$(RELEASE) upload --user kevinburke --repo differ --tag $(version) --name differ-windows-amd64 --file releases/$(version)/differ-windows-amd64 || true
