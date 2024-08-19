# High Level Design

This document is work in progress.

## Conf

The [Conf](confmap.go) represents the raw configuration for a service (e.g. OpenTelemetry Collector).

## Provider

The [Provider](provider.go) provides configuration, and allows to watch/monitor for changes. Any `Provider`
has a `<scheme>` associated with it, and will provide configs for `configURI` that follow the "<scheme>:<opaque_data>" format.
This format is compatible with the URI definition (see [RFC 3986](https://datatracker.ietf.org/doc/html/rfc3986)).
The `<scheme>` MUST be always included in the `configURI`. The scheme for any `Provider` MUST be at least 2
characters long to avoid conflicting with a driver-letter identifier as specified in
[file URI syntax](https://datatracker.ietf.org/doc/html/rfc8089#section-2).

## Converter

The [Converter](converter.go) allows implementing conversion logic for the provided configuration. One of the most
common use-case is to migrate/transform the configuration after a backwards incompatible change.

## Resolver

The `Resolver` handles the use of multiple [Providers](#provider) and [Converters](#converter)
simplifying configuration parsing, monitoring for updates, and the overall life-cycle of the used config providers.
The `Resolver` provides two main functionalities: [Configuration Resolving](#configuration-resolving) and
[Watching for Updates](#watching-for-updates).

### Configuration Resolving

The `Resolver` receives as input a set of `Providers`, a list of `Converters`, and a list of configuration identifier
`configURI` that will be used to generate the resulting, or effective, configuration in the form of a `Conf`,
that can be used by code that is oblivious to the usage of `Providers` and `Converters`.

`Providers` are used to provide an entire configuration when the `configURI` is given directly to the `Resolver`,
or an individual value (partial configuration) when the `configURI` is embedded into the `Conf` as a values using
the syntax `${configURI}`.

**Limitation:** 
- When embedding a `${configURI}` the uri cannot contain dollar sign ("$") character unless it embeds another uri.
- The number of URIs is limited to 100.

```terminal
              Resolver                   Provider
   Resolve       │                          │
────────────────►│                          │
                 │                          │
              ┌─ │        Retrieve          │
              │  ├─────────────────────────►│
              │  │          Conf            │
              │  │◄─────────────────────────┤
  foreach     │  │                          │
  configURI   │  ├───┐                      │
              │  │   │Merge                 │
              │  │◄──┘                      │
              └─ │                          │
              ┌─ │        Retrieve          │
              │  ├─────────────────────────►│
              │  │    Partial Conf Value    │
              │  │◄─────────────────────────┤
  foreach     │  │                          │
  embedded    │  │                          │
  configURI   │  ├───┐                      │
              │  │   │Replace               │
              │  │◄──┘                      │
              └─ │                          │
                 │            Converter     │
              ┌─ │     Convert    │         │
              │  ├───────────────►│         │
    foreach   │  │                │         │
   Converter  │  │◄───────────────┤         │
              └─ │                          │
                 │                          │
◄────────────────┤                          │
```

The `Resolve` method proceeds in the following steps:

1. Start with an empty "result" of `Conf` type.
2. For each config URI retrieves individual configurations, and merges it into the "result".
3. For each embedded config URI retrieves individual value, and replaces it into the "result".
4. For each "Converter", call "Convert" for the "result".
5. Return the "result", aka effective, configuration.

### Watching for Updates
After the configuration was processed, the `Resolver` can be used as a single point to watch for updates in the
configuration retrieved via the `Provider` used to retrieve the “initial” configuration and to generate the “effective” one.

```terminal      
         Resolver              Provider
            │                     │
   Watch    │                     │
───────────►│                     │
            │                     │
            .                     .
            .                     .
            .                     .
            │      onChange       │
            │◄────────────────────┤
◄───────────┤                     │
```

The `Resolver` does that by passing an `onChange` func to each `Provider.Retrieve` call and capturing all watch events. 

## Troubleshooting

### Null Maps

Due to how our underlying merge library, [koanf](https://github.com/knadh/koanf), behaves, configuration resolution
will treat configuration such as 

```yaml
processors:
```

as null, which is a valid value. As a result if you have configuration `A`:

```yaml
receivers:
    nop:

processors:
    nop:

exporters:
    nop:

extensions:
    nop:

service:
    extensions: [nop]
    pipelines:
        traces:
            receivers: [nop]
            processors: [nop]
            exporters: [nop]
```

and configuration `B`:

```yaml
processors:
```

and do `./otelcorecol --config A.yaml --config B.yaml`

The result will be an error:

```
Error: invalid configuration: service::pipelines::traces: references processor "nop" which is not configured
2024/06/10 14:37:14 collector server run finished with error: invalid configuration: service::pipelines::traces: references processor "nop" which is not configured
```

This happens because configuration `B` sets `processors` to null, removing the `nop` processor defined in configuration `A`,
so the `nop` processor referenced in configuration `A`'s pipeline no longer exists.

This situation can be remedied 2 ways:
1. Use `{}` when you want to represent an empty map, such as `processors: {}` instead of `processors:`.
2. Omit configuration like `processors:` from your configuration.
