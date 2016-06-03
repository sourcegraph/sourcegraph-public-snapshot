# que-go

[![GoDoc](https://godoc.org/github.com/bgentry/que-go?status.svg)][godoc]

NOTE: This is a hacked up version of the excellent `github.com/bgentry/que-go`
package. I want to use the usual `github.com/lib/pq`, since I am in a large app
which already uses it (and introducing pgx will require also introducing
monitoring/ etc). Not proven to work yet!!

Que-go is a fully interoperable Golang port of [Chris Hanks][chanks]' [Ruby Que
queuing library][ruby-que] for PostgreSQL. Que uses PostgreSQL's advisory locks
for speed and reliability.

Because que-go is an interoperable port of Que, you can enqueue jobs in Ruby
(i.e. from a Rails app) and write your workers in Go. Or if you have a limited
set of jobs that you want to write in Go, you can leave most of your workers in
Ruby and just add a few Go workers on a different queue name. Or you can just
write everything in Go :)

## pgx PostgreSQL driver

This package uses the [pgx][pgx] Go PostgreSQL driver rather than the more
popular [pq][pq]. Because Que uses session-level advisory locks, we have to hold
the same connection throughout the process of getting a job, working it,
deleting it, and removing the lock.

Pq and the built-in database/sql interfaces do not offer this functionality, so
we'd have to implement our own connection pool. Fortunately, pgx already has a
perfectly usable one built for us. Even better, it offers better performance
than pq due largely to its use of binary encoding.

Please see the [godocs][godoc] for more info and examples.

[godoc]: https://godoc.org/github.com/bgentry/que-go
[chanks]: https://github.com/chanks
[ruby-que]: https://github.com/chanks/que
[pgx]: https://github.com/jackc/pgx
[pq]: https://github.com/lib/pq
