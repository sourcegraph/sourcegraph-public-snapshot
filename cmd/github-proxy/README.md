# github-proxy

Proxies all requests to github.com to keep track of rate limits and prevent triggering abuse mechanisms.

There is only one replica running in production. However, we can have multiple replicas to increase our rate limits (rate limit is per IP).
