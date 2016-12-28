# JavaScript and TypeScript langserver

To build & update the JavaScript/TypeScript langserver:

```
./build.sh
```

To run the JavaScript/TypeScript langserver with a Sourcegraph development server:

```
export LANGSERVER_JAVASCRIPT=tcp://localhost:2088
export LANGSERVER_TYPESCRIPT=tcp://localhost:2088
make serve-dev
```

And in another terminal, start the JavaScript/TypeScript langserver process:

```
node ./langserver/lib/language-server.js -p 2088 --strict
```
