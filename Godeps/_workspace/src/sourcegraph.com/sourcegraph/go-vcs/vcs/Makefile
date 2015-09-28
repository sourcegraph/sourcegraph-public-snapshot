.PHONY: benchmark.txt

benchmark.txt:
	go test -test.bench='.*' -test.benchmem > $@ 2>&1
	cat $@
