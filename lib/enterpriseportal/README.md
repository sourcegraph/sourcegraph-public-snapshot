# Enterprise Portal services

`lib/enterpriseportal` defines the gRPC services implemented by Enterprise Portal. Core functionality are defined in `core/v1`, with extensions defined as separate services implemented by Enterprise Portal, such as `cody/v1`.

```sh
sg gen buf lib/enterpriseportal/subscriptions/v1/buf.gen.yaml lib/enterpriseportal/codygateway/v1/buf.gen.yaml
```

> **DRAFT STATE** - see [RFC 885](https://docs.google.com/document/d/1tiaW1IVKm_YSSYhH-z7Q8sv4HSO_YJ_Uu6eYDjX7uU4/edit#heading=h.tdaxc5h34u7q).
