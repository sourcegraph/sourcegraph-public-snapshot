#!/bin/sh
template_dir="${NGINX_ENVSUBST_TEMPLATE_DIR:-/etc/nginx/templates}"
suffix="${NGINX_ENVSUBST_TEMPLATE_SUFFIX:-.template}"
output_dir="${NGINX_ENVSUBST_OUTPUT_DIR:-/etc/nginx/conf.d}"
filter="${NGINX_ENVSUBST_FILTER:-}"
# shellcheck disable=SC2046
defined_envs=$(printf "\${%s} " $(awk "END { for (name in ENVIRON) { print ( name ~ /${filter}/ ) ? name : \"\" } }" </dev/null))

for template in /etc/nginx/templates/*.template; do
  relative_path="${template#"$template_dir/"}"
  output_path="$output_dir/${relative_path%"$suffix"}"
  subdir=$(dirname "$relative_path")
  mkdir -p "$output_dir/$subdir"
  echo "Processing $template -> $output_path"
  envsubst "$defined_envs" <"$template" >"$output_path"
done

exec "$@"
