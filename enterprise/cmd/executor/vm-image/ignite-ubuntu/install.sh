#!/usr/bin/env bash
set -ex -o nounset -o pipefail

## Creates a systemd service that starts before the docker daemon and checks
## for the existence of a /docker-registry.txt file. If present, it will configure
## the daemon before it returns, so that we don't need to restart it after boot
## when the config changed.
function setup_registry_overwrite() {
  cat <<EOF >/etc/systemd/system/registry_overwrite.service
[Unit]
Description=Configure registry mirror on startup, if defined.
Before=docker.service

[Service]
ExecStart=/usr/local/bin/registry_overwrite.sh

[Install]
WantedBy=docker.service
EOF

  cat <<EOF >/usr/local/bin/registry_overwrite.sh
#!/usr/bin/env bash

if [[ -f "/docker-registry.txt" ]]; then
  mkdir -p /etc/docker
  echo 'Found /docker-registry.txt file, configuring docker daemon..'
  DOCKER_REGISTRY_MIRROR="\$(cat /docker-registry.txt)"
  if [[ -f "/etc/docker/daemon.json" ]]; then
      cat /etc/docker/daemon.json | jq ".\"registry-mirrors\" |= [\"\${DOCKER_REGISTRY_MIRROR}\"]" > /etc/docker/daemon.json.tmp
      mv /etc/docker/daemon.json.tmp /etc/docker/daemon.json
  else
      echo "{\"registry-mirrors\": [\"\${DOCKER_REGISTRY_MIRROR}\"]}" > /etc/docker/daemon.json
  fi
fi

# Proceed with startup.
EOF

  # Ensure systemd can execute the script.
  chmod +x /usr/local/bin/registry_overwrite.sh

  systemctl enable registry_overwrite
}

setup_registry_overwrite
