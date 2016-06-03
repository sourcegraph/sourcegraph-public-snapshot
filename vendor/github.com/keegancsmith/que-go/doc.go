/*
Package que-go is a fully interoperable Golang port of Chris Hanks' Ruby Que
queueing library for PostgreSQL. Que uses PostgreSQL's advisory locks
for speed and reliability. See the original Que documentation for more details:
https://github.com/chanks/que

Because que-go is an interoperable port of Que, you can enqueue jobs in Ruby
(i.e. from a Rails app) and write your workers in Go. Or if you have a limited
set of jobs that you want to write in Go, you can leave most of your workers in
Ruby and just add a few Go workers on a different queue name.

PostgreSQL Driver pgx

Instead of using database/sql and the more popular pq PostgreSQL driver, this
package uses the pgx driver: https://github.com/jackc/pgx

Because Que uses session-level advisory locks, we have to hold the same
connection throughout the process of getting a job, working it, deleting it, and
removing the lock.

Pq and the built-in database/sql interfaces do not offer this functionality, so
we'd have to implement our own connection pool. Fortunately, pgx already has a
perfectly usable one built for us. Even better, it offers better performance
than pq due largely to its use of binary encoding.

Prepared Statements

que-go relies on prepared statements for performance. As of now these have to
be initialized manually on your connection pool like so:

    pgxpool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
        ConnConfig:   pgxcfg,
        AfterConnect: que.PrepareStatements,
    })

If you have suggestions on how to cleanly do this automatically, please open an
issue!

Usage

Here is a complete example showing worker setup and two jobs enqueued, one with a delay:

    type printNameArgs struct {
        Name string
    }

    printName := func(j *que.Job) error {
        var args printNameArgs
        if err := json.Unmarshal(j.Args, &args); err != nil {
            return err
        }
        fmt.Printf("Hello %s!\n", args.Name)
        return nil
    }

    pgxcfg, err := pgx.ParseURI(os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal(err)
    }

    pgxpool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
        ConnConfig:   pgxcfg,
        AfterConnect: que.PrepareStatements,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer pgxpool.Close()

    qc := que.NewClient(pgxpool)
    wm := que.WorkMap{
        "PrintName": printName,
    }
    workers := que.NewWorkerPool(qc, wm, 2) // create a pool w/ 2 workers
    go workers.Start() // work jobs in another goroutine

    args, err := json.Marshal(printNameArgs{Name: "bgentry"})
    if err != nil {
        log.Fatal(err)
    }

    j := &que.Job{
        Type:  "PrintName",
        Args:  args,
    }
    if err := qc.Enqueue(j); err != nil {
        log.Fatal(err)
    }

    j := &que.Job{
        Type:  "PrintName",
        RunAt: time.Now().UTC().Add(30 * time.Second), // delay 30 seconds
        Args:  args,
    }
    if err := qc.Enqueue(j); err != nil {
        log.Fatal(err)
    }

    time.Sleep(35 * time.Second) // wait for while

    workers.Shutdown()

*/
package que
