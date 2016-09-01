set -ex
# This script is intended to be run from the repo root.

# generates app/sitemap/sitemap_repo_top1k.xml
bash app/sitemap/list-top1k-repos-by-stars.sh go \
	| go run app/sitemap/make_sitemap.go \
	| gzip \
	> app/sitemap/sitemap_repo_top1k.xml.gz
