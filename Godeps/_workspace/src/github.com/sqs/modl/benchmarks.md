Current benchmarks:

    PASS
    BenchmarkNativeCrud	     100	  22605756 ns/op	    5402 B/op	     107 allocs/op
    BenchmarkModlCrud	     100	  21107020 ns/op	    7329 B/op	     153 allocs/op
    ok  	_/mnt/omocha/jmoiron/dev/go/modl	5.975s
    PASS
    BenchmarkNativeCrud	     100	  13153419 ns/op	    8822 B/op	     334 allocs/op
    BenchmarkModlCrud	     100	  11676837 ns/op	   10577 B/op	     381 allocs/op
    ok  	_/mnt/omocha/jmoiron/dev/go/modl	3.813s
    PASS
    BenchmarkNativeCrud	    5000	    281356 ns/op	    1455 B/op	      51 allocs/op
    BenchmarkModlCrud	    5000	    313549 ns/op	    2713 B/op	      90 allocs/op
    ok  	_/mnt/omocha/jmoiron/dev/go/modl	3.136s

(In order: MySQL, PostgreSQL, SQLite)

