#!/bin/sh
/usr/bin/cadvisor \
  -enable_metrics=cpu,memory,disk,network \
  -logtostderr \
  -port=48080 \
  -docker_only \
  -housekeeping_interval=10s \
  -max_housekeeping_interval=15s \
  -event_storage_event_limit=default=0 \
  -v=3 \
  -event_storage_age_limit=default=0 \
  -containerd=/var/run/containerd/containerd.sock
