#!/usr/bin/env bash

set -eu

if [ -z "$VERSION" ]; then
  echo "❌ Need \$VERSION to be set to promote images"
  exit 1
fi

if [ "$#" -lt 1 ]; then
  echo "❌ Usage: $0 gitserver blobstore <image-name-without-registry...> ..."
  exit 1
fi

# We're transitioning to GAR because of DockerHub new rate limiting affecting GCP
# See https://github.com/sourcegraph/sourcegraph/issues/61696
# Set IFS to space and read into an array
IFS=' ' read -r -a PROMOTION_REGISTRIES <<< "$ADDITIONAL_PROD_REGISTRIES"
PROMOTION_REGISTRIES=("$PUBLIC_REGISTRY" "${PROMOTION_REGISTRIES[@]}" )

if [ ! -e "./annotations" ]; then
    mkdir ./annotations
fi
echo -e "## Release: image promotions" > ./annotations/image_promotions.md
echo -e "\n| Name | From | To |\n|---|---|---|" >> ./annotations/image_promotions.md
for name in "${@:1}"; do
  echo "--- Copying ${name} from private registry to public registries"

  # Pull the internal release
  docker pull "${INTERNAL_REGISTRY}/${name}:${VERSION}"

  # Push it on the classic public registry (DockerHub)
  pushed=""
  for registry in "${PROMOTION_REGISTRIES[@]}"; do
    target="${registry}/${name}:${VERSION}"
    docker tag "${INTERNAL_REGISTRY}/${name}:${VERSION}" "$target"
    docker push "$target"
    pushed="$pushed \`$target\`"
  done

  echo -e "| ${name} | \`${INTERNAL_REGISTRY}/${name}:${VERSION}\` | ${pushed} |" >>./annotations/image_promotions.md
done
