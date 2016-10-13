# typescript langserver

To build & update the typescript langserver:

```
make
```

To run the typescript langserver with a Sourcegraph development server:

```
export LANGSERVER_TYPESCRIPT=tcp://localhost:2088
make serve-dev
```

And in another terminal, start the typescript langserver process:

```
node ./langserver/build/language-server.js -p 2088 --strict
```