### GQL2TS

[![Build Status](https://travis-ci.org/avantcredit/gql2ts.svg?branch=refactor_with_tests)](https://travis-ci.org/avantcredit/gql2ts)

```shell
npm install -g gql2ts
```


```
Usage: gql2ts [options] <schema.json>

Options:

  -h, --help                         output usage information
  -V, --version                      output the version number
  -o --output-file [outputFile]      name for ouput file, defaults to graphqlInterfaces.d.ts
  -n --namespace [namespace]         name for the namespace, defaults to "GQL"
  -i --ignored-types <ignoredTypes>  names of types to ignore (comma delimited)
```

### Examples

#### With Default Options
```shell
gql2ts schema.json
```


#### With Optional Options
```shell
gql2ts -n Avant -i BadInterface,BadType,BadUnion -o avant-gql.d.ts schema.json
```
