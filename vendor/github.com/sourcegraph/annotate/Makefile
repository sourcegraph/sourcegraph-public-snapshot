.PHONY: benchmark.txt

benchmark.txt:
	go test -test.run='^$$' -test.bench='.*' -test.benchmem > $@ 2>&1
	cat $@

profile:
	go test -test.run='^$$' -test.bench='.*' -test.cpuprofile=/tmp/annotate.prof
	go tool pprof ./annotate.test /tmp/annotate.prof
