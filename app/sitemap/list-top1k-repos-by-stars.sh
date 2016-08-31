set -ex
BASE_URL="https://api.github.com/search/repositories"

# $1: language
# $2: output file

for i in {1..10}; do
	curl -XGET "$BASE_URL?q=language:$1&sort=stars&per_page=100&page=$i" \
		| grep "git_url\|stargazers_count"  \
		| sed 's/^ *\"[a-z_]*\": \([^,]*\),$/\1/' \
		| sed 'N;s/\n/ /' \
		| sed 's/\.git\"/\/-\/land/' \
		| sed 's/\"git:\/\//https:\/\/sourcegraph.com\//' >> $2;
done
