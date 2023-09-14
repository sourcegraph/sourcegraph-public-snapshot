# Sourcegraph sitemap generator

This tool is ran offline to generate the sitemap files served at https://sourcegraph.com/sitemap.xml

To run it:

```sh
export SRC_ACCESS_TOKEN=...
./run.sh
```

Once ran, it will output some stats as well as generate the sitemap.xml files to `sitemap/`. You should then upload them:

```sh
gsutil cp -r sitemap/ gs://sitemap-sourcegraph-com
gsutil ls 'gs://sitemap-sourcegraph-com/*.gz' | xargs gsutil setmeta -h 'Content-Encoding:gzip'
```
