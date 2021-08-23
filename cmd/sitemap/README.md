# Sourcegraph sitemap generator

This tool is ran offline to generate the sitemap files served at https://sourcegraph.com/sitemap.xml

To run it:

```sh
go build -o sitemap-generator ./cmd/sitemap && ./sitemap-generator
```

