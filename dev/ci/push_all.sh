#!/usr/bin/env bash

set -eu

echo "~~~ :aspect: :stethoscope: Agent Health check"
/etc/aspect/workflows/bin/agent_health_check

aspectRC="/tmp/aspect-generated.bazelrc"
rosetta bazelrc >"$aspectRC"
bazelrc=(--bazelrc="$aspectRC" --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc)

function preview_tags() {
  IFS=' ' read -r -a registries <<<"$1"
  IFS=' ' read -r -a tags <<<"$2"

  for tag in "${tags[@]}"; do
    for registry in "${registries[@]}"; do
      echo -e "\t ${registry}/\$IMAGE:${tag}"
    done
  done
}

# Append to annotations which image was pushed and with which tags.
# Because this is meant to be executed by parallel, meaning we write commands
# to a jobfile, this echoes the command to post the annotation instead of actually
# doing it.
function echo_append_annotation() {
  repository="$1"
  registry="$2"
  IFS=' ' read -r -a tag_args <<<"$3"
  formatted_tags=""

  for arg in "${tag_args[@]}"; do
    if [ "$arg" != "--tag" ]; then
      if [ "$formatted_tags" == "" ]; then
        # Do not insert a comma for the first element
        formatted_tags="<code>$arg</code>"
      else
        formatted_tags="${formatted_tags}, <code>$arg</code>"
      fi
    fi
  done

  raw="<tr><td>${repository}</td><td><code>${registry}</code></td><td>${formatted_tags}</td></tr>"
  echo "echo -e '${raw}' >>./annotations/pushed_images.md"
}

function create_push_command() {
  IFS=' ' read -r -a registries <<<"$1"
  repository="$2"
  target="$3"
  tags_args="$4"

  # TODO(JH): https://github.com/sourcegraph/sourcegraph/issues/58442
  if [[ "$target" == "//docker-images/syntax-highlighter:scip-ctags_candidate_push" ]]; then
    repository="scip-ctags"
  fi

  for registry in "${registries[@]}"; do
    # This biblical bash string replaces running `bazel run` on each oci_push target, to avoid the (temporary) bazel server lock
    # that is unnecessary for us due to building all the targets beforehand, allowing the maximum possible concurrency. It is similar
    # to the script that is emitted by `bazel run --run_script=out.sh <target>`, but without the need to wait for the server to be unlocked
    # in the same way as just running `bazel run`.
    # It makes the following assumptions:
    # - the executable script for the oci_push target is named push_<target name>.sh
    # - the target is built and exists in the bazel bindir (we do this with a bazel build below)
    # - runfiles are always adajcent to the executable script

    # echo to /dev/null is for the final output in Buildkite
    echo "echo $target >/dev/null && \
      pushd $(realpath bazel-bin)$(echo "${target}.sh.runfiles/__main__" | sed 's/:/\/push_/') && \
      $(realpath bazel-bin)$(echo "${target}.sh" | sed 's/:/\/push_/') $tags_args --repository ${registry}/${repository} && \
      popd && $(echo_append_annotation "$repository" "$registry" "${tags_args[@]}")"
  done
}

dev_registries=(
  "$DEV_REGISTRY"
)

prod_registries=(
  "$PROD_REGISTRY"
)

if [ -n "${ADDITIONAL_PROD_REGISTRIES}" ]; then
  IFS=' ' read -r -a registries <<<"$ADDITIONAL_PROD_REGISTRIES"
  prod_registries+=("${registries[@]}")
fi

date_fragment="$(date +%Y-%m-%d)"

dev_tags=(
  "${BUILDKITE_COMMIT:0:12}"
  "${BUILDKITE_COMMIT:0:12}_${date_fragment}"
  "${BUILDKITE_COMMIT:0:12}_${BUILDKITE_BUILD_NUMBER}"
)
prod_tags=(
  "${PUSH_VERSION}"
)

CANDIDATE_ONLY=${CANDIDATE_ONLY:-""}

push_prod=false

# If we're doing an internal release, we need to push to the prod registry too.
# TODO(rfc795) this should be more granular than this, we're abit abusing the idea of the prod registry here.
if [ "${RELEASE_INTERNAL:-}" == "true" ]; then
  push_prod=true
elif [[ "$BUILDKITE_BRANCH" =~ ^main$ ]] || [[ "$BUILDKITE_BRANCH" =~ ^docker-images-candidates-notest/.* ]]; then
  dev_tags+=("insiders")
  prod_tags+=("insiders")
  push_prod=true
elif [[ "$BUILDKITE_BRANCH" =~ ^main-dry-run/.* ]]; then
  # We only push on internal registries on a main-dry-run.
  dev_tags+=("insiders")
  prod_tags+=("insiders")
  push_prod=false
elif [[ "$BUILDKITE_BRANCH" =~ ^docker-images/.* ]]; then
  # We only push on internal registries on a main-dry-run.
  dev_tags+=("insiders")
  prod_tags+=("insiders")
  push_prod=true
elif [[ "$BUILDKITE_BRANCH" =~ ^[0-9]+\.[0-9]+$ ]]; then
  # All release branch builds must be published to prod tags to support
  # format introduced by https://github.com/sourcegraph/sourcegraph/pull/48050
  # by release branch deployments.
  push_prod=true
elif [[ "$BUILDKITE_BRANCH" =~ ^[0-9]+\.[0-9]+\.(x|[0-9]+)$ ]]; then
  # Patch release builds only need to be pushed to internal registries.
  push_prod=false
  dev_tags+=("$BUILDKITE_BRANCH-insiders")
elif [[ "$BUILDKITE_TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(\-rc\.[0-9]+)?$ ]]; then
  # ok: v5.1.0
  # ok: v5.1.0-rc.5
  # no: v5.1.0-beta.1
  # no: v5.1.0-rc5
  dev_tags+=("${BUILDKITE_TAG:1}")
  prod_tags+=("${BUILDKITE_TAG:1}")
  push_prod=true
fi

# If we're building ephemeral cloud images, we don't push to prod but we need to prod version as tag
if [ "${CLOUD_EPHEMERAL:-}" == "true" ]; then
  dev_tags=("${PUSH_VERSION}")
  push_prod=false
fi

# If CANDIDATE_ONLY is set, only push the candidate tag to the dev repo
if [ -n "$CANDIDATE_ONLY" ]; then
  dev_tags=("${BUILDKITE_COMMIT}_${BUILDKITE_BUILD_NUMBER}_candidate")
  push_prod=false
fi

# Posting the preamble for image pushes.
echo -e "### ${BUILDKITE_LABEL}" >./annotations/pushed_images.md
echo -e "<details><summary>Click to expand table</summary><table>\n" >>./annotations/pushed_images.md
echo -e "<tr><th>Name</th><th>Registry</th><th>Tags</th></tr>\n" >>./annotations/pushed_images.md

preview_tags "${dev_registries[*]}" "${dev_tags[*]}"
if $push_prod; then
  preview_tags "${prod_registries[*]}" "${prod_tags[*]}"
fi

dev_tags_args=""
for t in "${dev_tags[@]}"; do
  dev_tags_args="$dev_tags_args --tag ${t}"
done
prod_tags_args=""
if $push_prod; then
  for t in "${prod_tags[@]}"; do
    prod_tags_args="$prod_tags_args --tag ${t}"
  done
fi

images=$(bazel "${bazelrc[@]}" query 'kind("oci_push rule", //...)')

echo "--- :bazel: Building all oci_push targets"

# shellcheck disable=SC2086
bazel "${bazelrc[@]}" build \
  --announce_rc \
  --profile=bazel-profile.gz \
  --experimental_execution_log_compact_file=execution_log.zstd \
  --stamp --workspace_status_command=./dev/bazel_stamp_vars.sh \
  --build_event_binary_file=build_event_log.bin \
  --build_event_binary_file_path_conversion=false \
  --build_event_binary_file_upload_mode=wait_for_upload_complete \
  --build_event_publish_all_actions=true \
  --remote_download_outputs=toplevel \
  ${images}

echo "--- :bash: Generating jobfile - started"

job_file=$(mktemp)
# shellcheck disable=SC2064
trap "rm -rf $job_file" EXIT

# shellcheck disable=SC2068
for target in ${images[@]}; do
  [[ "$target" =~ ([A-Za-z0-9_.-]+): ]]
  name="${BASH_REMATCH[1]}"
  # Append push commands for dev registries
  create_push_command "${dev_registries[*]}" "$name" "$target" "$dev_tags_args" >>"$job_file"
  # Append push commands for prod registries
  if $push_prod; then
    create_push_command "${prod_registries[*]}" "$name" "$target" "$prod_tags_args" >>"$job_file"
  fi
done

echo "--- :bash: Generating jobfile - done"
cat "$job_file"

echo "--- :bazel::docker: Pushing images..."
log_file=$(mktemp)
# shellcheck disable=SC2064
trap "rm -rf $log_file" EXIT

# See dev/ci/internal/ci/images_operations.go
JOBS="${PUSH_CONCURRENT_JOBS:-4}"

parallel --jobs="$JOBS" --line-buffer --joblog "$log_file" -v <"$job_file"

# Pretty print the output from gnu parallel
while read -r line; do
  # Skip the first line (header)
  if [[ "$line" != Seq* ]]; then
    cmd="$(echo "$line" | cut -f9)"
    [[ "$cmd" =~ (\/\/[^ ]+) ]]
    target="${BASH_REMATCH[1]}"
    exitcode="$(echo "$line" | cut -f7)"
    duration="$(echo "$line" | cut -f4 | tr -d "[:blank:]")"
    if [ "$exitcode" == "0" ]; then
      echo "--- :docker::arrow_heading_up: $target ${duration}s :white_check_mark:"
    else
      echo "--- :docker::arrow_heading_up: $target ${duration}s: failed with $exitcode) :red_circle:"
    fi
  fi
done <"$log_file"

echo -e "</table></details>" >>./annotations/pushed_images.md

echo "--- :bazel::docker: detailed summary"
cat "$log_file"
