if [ "$#" -ne 2 ]; then
    echo "Illegal number of parameters. Please read ./dev/server/README.md"
    exit -1
fi

lang=$1
version=$2
docker pull us.gcr.io/sourcegraph-dev/xlang-$lang:$version
docker tag us.gcr.io/sourcegraph-dev/xlang-$lang:$version sourcegraph/codeintel-$lang:$version
docker tag us.gcr.io/sourcegraph-dev/xlang-$lang:$version sourcegraph/codeintel-$lang:latest
docker push sourcegraph/codeintel-$lang:$version
docker push sourcegraph/codeintel-$lang:latest