.PHONY: benchmark.txt

benchmark.txt:
	go test -test.run='^$$' -test.bench='.*' -test.benchmem > $@ 2>&1
	cat $@
