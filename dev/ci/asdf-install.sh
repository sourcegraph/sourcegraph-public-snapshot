#!/usr/bin/env bash

set -e

# ASDF setup that either does a simple install, or pulls it from cache, geared towards
# usage in CI.
# In most cases you should not need to call this script directly.
if [[ ! "$BUILDKITE" == "true" ]]; then
  # Not-in-buildkite simple install.
  echo "asdf install"
  asdf install
  echo "done installing"
  # We can't use exit 0 here, it would prevent the variables to be exported (that's a particular buildkite hook peculiarity).
else
  # We need awscli to use asdf cache
  echo "asdf install from cache"
  asdf install awscli
  echo "done installing awscli"

  # set the buildkite cache access keys
  AWS_CONFIG_DIR_PATH="/buildkite/.aws"
  mkdir -p "$AWS_CONFIG_DIR_PATH"
  AWS_CONFIG_FILE="$AWS_CONFIG_DIR_PATH/config"
  export AWS_CONFIG_FILE
  AWS_SHARED_CREDENTIALS_FILE="/buildkite/.aws/credentials"
  export AWS_SHARED_CREDENTIALS_FILE
  aws configure set aws_access_key_id "$BUILDKITE_HMAC_KEY" --profile buildkite
  aws configure set aws_secret_access_key "$BUILDKITE_HMAC_SECRET" --profile buildkite

  asdf_checksum=$(sha1sum .tool-versions | awk '{print $1}')
  cache_file="cache-asdf-$asdf_checksum.tar.gz"
  cache_key="$BUILDKITE_ORGANIZATION_SLUG/$BUILDKITE_PIPELINE_NAME/$cache_file"

  echo -e "ASDF 🔍 Locating cache: $cache_key"
  if aws s3api head-object --bucket "sourcegraph_buildkite_cache" --profile buildkite --endpoint-url 'https://storage.googleapis.com' --region "us-central1" --key "$cache_key"; then
    echo -e "ASDF 🔥 Cache hit: $cache_key"
    aws s3 cp --profile buildkite --no-progress --endpoint-url 'https://storage.googleapis.com' --region "us-central1" "s3://sourcegraph_buildkite_cache/$cache_key" "$HOME/"
    pushd "$HOME" || exit
    rm -rf .asdf
    bsdtar xzf "$cache_file"
    popd || exit
  else
    echo -e "ASDF 🚨 Cache miss: $cache_key"
    echo "~~~ fresh install of all asdf tool versions"
    asdf install
    echo "~~~ cache asdf installation"
    pushd "$HOME" || exit
    bsdtar cfz "$cache_file" .asdf
    popd || exit
    aws s3 cp --profile buildkite --no-progress --endpoint-url 'https://storage.googleapis.com' --region "us-central1" "$HOME/$cache_file" "s3://sourcegraph_buildkite_cache/$cache_key"
  fi

  unset AWS_SHARED_CREDENTIALS_FILE
  unset AWS_CONFIG_FILE
fi

asdf reshim
