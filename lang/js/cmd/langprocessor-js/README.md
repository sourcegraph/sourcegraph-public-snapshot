# JavaScript Language Processor Proxy Server

## Prerequisites

 - [NodeJS](https://nodejs.org/en/download/)
 - TypeScript (`npm install -g typescript`)

## Running

Build and release docker image:

```bash
./build.sh

...

Successfully built e71e5e9e9ebf
```

Run and expose the port:

```bash
docker run -p 4145:4145 -it e71e5e9e9ebf
```
