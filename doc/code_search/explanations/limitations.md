# Limitations

## Timeouts

There are two sources of timeouts Sourcegraph query:

- A timeout in the HTTP load balancer in front of Sourcegraph (nginx/ELB/Cloudflare/etc). Your admin will likely need to increase timeouts for Sourcegraph endpoints. In particular the `.api/search/stream` path. This uses [SSE](https://en.wikipedia.org/wiki/Server-sent_events), so your reverse proxy may have specific support for these requests.
- A maximum timeout enforced by Sourcegraph. Your admin may need to increase the site configuration value (default 60s) with the following setting:

```json
"search.limits": {
    "maxTimeoutSeconds": 60,
  },
```

## Non-indexed backends

A search is unindexed if you are searching non-indexed branches or using diff/commit search. Using a non-indexed backend and searching all code in a large instance can take 10min+. This is likely much higher than any configured timeouts. See the [Timeouts](#timeouts) section on how to configure this use case. See [Search-Jobs](../how-to/search-jobs.md) to run a search in the background over all code.

Currently, our non-indexed backends do not use the same scheduling logic as indexed backends. This means concurrent slow non-indexed searches will impact resources of interactive searches.

An unindexed search can under-report result counts. This is due to limits on the number of results reported per file. See [#18298](https://github.com/sourcegraph/sourcegraph/issues/18298).
