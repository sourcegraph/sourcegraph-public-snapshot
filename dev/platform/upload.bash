#!/usr/bin/env bash

set -ex
unset CDPATH
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

HOST=${HOST-ec2-13-56-175-167.us-west-1.compute.amazonaws.com}
HOSTDIR=cxp

TMPDIR=$(mktemp -d)
function finish {
    rm -rf "$TMPDIR"
}
trap finish EXIT

mkdir -p "$TMPDIR"/systemd

declare -A extensions
extensions[cx-blame]="4000"
extensions[cx-codecov]="4001"
extensions[cx-lightstep]="4002"
extensions[cx-line-age]="4003"
extensions[cx-logdna]="4005"
extensions[cx-sample-line-colors]="4006"
extensions[cx-sample-lsp]="4007"

GOBIN="$TMPDIR"/bin GOOS=linux go install -ldflags="-s -w" ./cxp/cmd/...

for prog in "${!extensions[@]}"; do
    cat <<EOF > "$TMPDIR"/systemd/"$prog".service
[Unit]
Description=${prog}

[Service]
Environment='CX_MODE=tcp' 'CX_ADDR=:${extensions[$prog]}' 'SOURCEGRAPH_CONFIG_FILE=/dev/null'
ExecStart=/home/ec2-user/${HOSTDIR}/bin/${prog}
TimeoutSec=30
Restart=on-failure
RestartSec=2
StartLimitInterval=350
StartLimitBurst=10

[Install]
WantedBy=multi-user.target
EOF
done

ssh "$HOST" mkdir -p "$HOSTDIR"
rsync --progress -avz "$TMPDIR"/ "$HOST":"$HOSTDIR"

for prog in "${!extensions[@]}"; do
    echo "cp ${HOSTDIR}/systemd/${prog}.service /etc/systemd/system && systemctl reset-failed ${prog} && systemctl start ${prog} && systemctl reload-or-restart ${prog} && systemctl enable ${prog}" | ssh "$HOST" sudo bash -
done
