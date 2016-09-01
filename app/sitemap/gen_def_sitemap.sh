set -ex
DO NOT RUN THIS SCRIPT unless you really mean to.
# It might cause tremendous load on the production DB.
# If you do intend to run it, comment the second line.

# This script is intended to be run from the repo root.

# generates app/sitemap/sitemap_def_top4k.xml.gz
src list_top_defs --limit=4000 --endpoint="https://grpc.sourcegraph.com" \
	| go run app/sitemap/make_sitemap.go \
	| gzip \
	> app/sitemap/sitemap_def_top4k.xml.gz
