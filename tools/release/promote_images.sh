#!/usr/bin/env bash

set -eu

if [ "$VERSION" = "" ]; then
  echo "❌ Need \$VERSION to be set to promote images"
  exit 1
fi

if [ "$#" -lt 1 ]; then
  echo "❌ Usage: $0 gitserver blobstore <image-name-without-registry...> ..."
  exit 1
fi

echo -e "## Release: image promotions" > ./annotations/image_promotions.md
echo -e "\n| Name | From | To |\n|---|---|---|" >> ./annotations/image_promotions.md
for name in "${@:1}"; do
  echo "--- Copying ${name} from private registry to public registries"

  # Pull the internal release
  docker pull "${INTERNAL_REGISTRY}/${name}:${VERSION}"

  # Push it on the classic public registry (DockerHub)
  docker tag "${INTERNAL_REGISTRY}/${name}:${VERSION}" "${PUBLIC_REGISTRY}/${name}:${VERSION}"
  docker push "${PUBLIC_REGISTRY}/${name}:${VERSION}"

  # We're transitioning to GAR because of DockerHub new rate limiting affecting GCP
  # See https://github.com/sourcegraph/sourcegraph/issues/61696
  docker tag "${INTERNAL_REGISTRY}/${name}:${VERSION}" "${ADDITIONAL_PROD_REGISTRY}/${name}:${VERSION}"
  docker push "${ADDITIONAL_PROD_REGISTRY}/${name}:${VERSION}"

  echo -e "| ${name} | \`${INTERNAL_REGISTRY}/${name}:${VERSION}\` | \`${PUBLIC_REGISTRY}/${name}:${VERSION}\` \`${ADDITIONAL_PROD_REGISTRY}/${name}:${VERSION}\` |" >>./annotations/image_promotions.md
done
