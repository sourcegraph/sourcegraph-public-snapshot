This directory is a place to put Grafana json dashboard specifications while running in local dev mode.
They will be picked up automatically and changes to them will be picked up automatically.

Once dashboards graduate to inclusion by default in a Sourcegraph installation they should migrate to
`docker-images/grafana/config/provisioning/dashboards/sourcegraph`.
