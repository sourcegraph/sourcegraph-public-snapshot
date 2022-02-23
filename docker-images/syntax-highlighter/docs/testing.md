# Running Tests

Currently, we use `insta` to run snapshot tests (and cargo as the test runner).

## Using `insta`

```
$ cargo install cargo-insta
```

Run tests with `$ cargo test`. If you have failures, you can do this cool thing:

```
$ cargo insta review
```

And that will lead you through any failures from the snapshot tests

(more to write here and links to add later)
