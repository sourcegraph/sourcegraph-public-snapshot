all: run

build: clean
	go build

clean:
	rm -f deployment-lag-notifier

run: clean build
	# Aggressive thresholds for local testing
	./deployment-lag-notifier -dry-run -num-commits 10 -allowed-age 10m

prod: clean build
	./deployment-lag-notifier -num-commits 20 -allowed-age 2h30m

test:
	go test -v
