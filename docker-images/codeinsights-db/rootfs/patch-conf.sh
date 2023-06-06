#!/usr/bin/env bash

# In Wolfi, unix_socket_directories defaults to /tmp. In previous Alpine images, this defaulted to /var/run/postgres.
# /tmp may not be writable, so any existing postgresql.conf configs that predate the Wolfi migration should be patched to update this setting.

CONFIG_DIR=${PGDATA:-/data/pgdata-12}

conf_file="$CONFIG_DIR/postgresql.conf"
new_socket_dir="/var/run/postgresql"

# Check if the parameter already exists in the file
if grep -q "^\s*unix_socket_directories" "$conf_file"; then
  echo "unix_socket_directories already exists in $conf_file"
else
  # Append the setting to the end of the file
  echo "unix_socket_directories = '$new_socket_dir'" >>"$conf_file"
  echo "Updated unix_socket_directories in $conf_file"
fi
