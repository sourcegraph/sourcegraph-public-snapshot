#!/bin/bash

if ! helm dependencies update helm/bootstrap; then
  echo "FATAL: Failed to update chart."
  #echo "BUG 🪳🐝 🦋 🪲 🐞 🐜 🦗: We allowed a broken chart!"
  exit 1
fi
helm package helm/bootstrap
gcloud compute scp --tunnel-through-iap ./*.tgz playground:/var/www/vhost/appliance/helm
rm ./*.tgz
gcloud compute ssh --tunnel-through-iap playground --command "helm repo index /var/www/vhost/appliance/helm"
