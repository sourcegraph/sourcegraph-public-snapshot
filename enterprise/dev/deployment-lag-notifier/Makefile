all: run

build: clean
	go build

clean:
	rm -f deployment-lag-notifier

run: clean build
	./deployment-lag-notifier -dry-run

prod: clean build
	./deployment-lag-notifier

test:
	go test -v
