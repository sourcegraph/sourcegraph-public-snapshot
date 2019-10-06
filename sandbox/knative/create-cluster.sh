#!/bin/bash

set -eux

export PROJECT=sqs-sandbox-knative
export CLUSTER_NAME=knative
export REGION=asia-east2
export CLUSTER_ZONE=asia-east2-a

gcloud config set core/project $PROJECT

gcloud services enable \
       cloudapis.googleapis.com \
       container.googleapis.com \
       containerregistry.googleapis.com

gcloud beta container clusters create $CLUSTER_NAME \
       --addons=HorizontalPodAutoscaling,HttpLoadBalancing,Istio \
       --machine-type=n1-standard-4 \
       --cluster-version=latest --zone=$CLUSTER_ZONE \
       --enable-stackdriver-kubernetes --enable-ip-alias \
       --enable-autoscaling --min-nodes=1 --max-nodes=10 \
       --enable-autorepair \
       --scopes cloud-platform \
       --metadata disable-legacy-endpoints=true \
       --no-enable-basic-auth \
       --no-issue-client-certificate

gcloud compute addresses create knative--external-ip --region $REGION
export EXTERNAL_IP=$(gcloud beta compute addresses describe knative--external-ip --region asia-east2 --format 'value(address)')

kubectl create clusterrolebinding cluster-admin-binding \
        --clusterrole=cluster-admin \
        --user=$(gcloud config get-value core/account)

# seems that this needs to run 2x, first time it gives an error about "Serving" in yml file
kubectl apply --selector knative.dev/crd-install=true \
        --filename https://github.com/knative/serving/releases/download/v0.9.0/serving.yaml \
        --filename https://github.com/knative/eventing/releases/download/v0.9.0/release.yaml \
        --filename https://github.com/knative/serving/releases/download/v0.9.0/monitoring.yaml

kubectl apply --filename https://github.com/knative/serving/releases/download/v0.9.0/serving.yaml \
        --filename https://github.com/knative/eventing/releases/download/v0.9.0/release.yaml \
        --filename https://github.com/knative/serving/releases/download/v0.9.0/monitoring.yaml

# ... other useful snippets ...

kubectl patch svc istio-ingressgateway --namespace istio-system --patch '{"spec": { "loadBalancerIP": "'${EXTERNAL_IP}'" }}'

# ...


# To delete:
#
# gcloud container clusters delete $CLUSTER_NAME --zone $CLUSTER_ZONE

