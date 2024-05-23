# Enterprise Portal services

`lib/enterpriseportal` defines the gRPC services implemented by Enterprise Portal. Core functionality are defined in `subscriptions/v1`, with extensions defined as separate services implemented by Enterprise Portal, such as `codyaccess/v1`.

To regenerate all relevant bindings:

```sh
sg gen buf \
  lib/enterpriseportal/subscriptions/v1/buf.gen.yaml \
  lib/enterpriseportal/codyaccess/v1/buf.gen.yaml
```

> **EVERYTHING HERE IS IN A DRAFT STATE** - see [RFC 885](https://docs.google.com/document/d/1tiaW1IVKm_YSSYhH-z7Q8sv4HSO_YJ_Uu6eYDjX7uU4/edit#heading=h.tdaxc5h34u7q).
