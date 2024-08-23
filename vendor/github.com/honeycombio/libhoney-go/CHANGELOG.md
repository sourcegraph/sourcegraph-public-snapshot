# libhoney Changelog

### Fixed

- Pass bytes.Reader to http.Request, clean up pooledReader (#159) | | [lizthegrey](https://github.com/lizthegrey)

## 1.15.7 2022-01-04

### Fixed

- Don't crash on stream aborts, always add content length (#156) | [lizthegrey](https://github.com/lizthegrey)

### Maintenance

- Add re-triage workflow (#155) | [vreynolds](https://github.com/vreynolds)
- Bump github.com/vmihailenco/msgpack/v5 from 5.3.4 to 5.3.5 (#149)
- Bump github.com/DataDog/zstd from 1.4.8 to 1.5.0 (#153)
- Bump github.com/klauspost/compress from 1.13.5 to 1.13.6 (#145)

## 1.15.6 2021-11-03

### Fixed

- Ensure valid JSON even when individual events in a batch can't be marshalled (#151)

### Maintenance

- empower apply-labels action to apply labels (#150)
- add min go version to readme (#147)
- update certs in old CI image (#148)
- ci: remove buildevents from nightly (#144)
- ci: secrets management (#142)

## 1.15.5 2021-09-27

### Fixed

- fix race condition on Honeycomb.Flush() (#140) | [@bfreis](https://github.com/bfreis)

### Maintenance

- Change maintenance badge to maintained (#138)
- Adds Stalebot (#141)
- Add issue and PR templates (#136)
- Add OSS lifecycle badge (#135)
- Add community health files (#134)
- Bump github.com/klauspost/compress from 1.12.3 to 1.13.5 (#130, #137)
- Bump github.com/vmihailenco/msgpack/v5 from 5.2.0 to 5.3.4 (#133)

## 1.15.4 2021-07-21

### Maintenance

- Upgrade msgpack from v4 to v5. (#127)

## 1.15.3 2021-06-02

### Improvements

- Add more context to batch response parsing error (#116)

### Maintenance

- Add go 1.15 & 1.16 to the testing matrix (#114, #119)

## 1.15.2 2021-01-22

NOTE: v1.15.1 may cause update warnings due to checksum error, please use v1.15.2 instead.

### Maintenance

- Add Github action to manage project labels (#110)
- Automate the creation of draft releases when project is tagged (#109)

## 1.15.1 2021-01-14

### Improvements

- Fix data race on dynFields length in Builder.Clone (#72)

### Maintenance

- Update dependencies
- github.com/klauspost/compress from 1.11.2 to 1.11.4 (#105, #106)

## 1.15.0 2020-11-10

- Mask writekey when printing events (#103)

## 1.14.1 2020-9-24

- Add .editorconfig to help provide consistent IDE styling (#99)

## 1.14.0 2020-09-01

- Documentation - document potential failures if pendingWorkCapacity not specified
- Documentation - use Deprecated tags for deprecated fields
- Log when event batch is rejected with an invalid API key
- Dependency bump (compress)

## 1.13.0 2020-08-21

- This release includes a change by @apechimp that makes Flush thread-safe (#80)
- Update dependencies
- Have a more obvious default statsd prefix (libhoney)
