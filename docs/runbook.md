## Access Sourcegraph.com database

```bash
gcloud container clusters get-credentials main-cluster-5
kubectl --namespace=prod exec -it $(kubectl --namespace=prod get pod -l app=pgsql -o jsonpath='{.items[0].metadata.name}') -- psql -U sg
```

## Delete repo(s) from cluster

```sql
begin;
delete from global_dep where repo_id in (select id from repo where uri not like '%/sourcegraph/%');
delete from pkgs where repo_id in (select id from repo where uri not like '%/sourcegraph/%');
delete from repo where uri not like '%/sourcegraph/%'
commit;
```

## Deploy dogfood cluster helm chart

1. Update image tags: https://sourcegraph.sgdev.org/github.com/sourcegraph/infrastructure@365d1fc3d3b6b2394d415ed7ff96e35c7156ab12/-/blob/kubernetes/cmd/sourcegraph-server-gen/generate.go#L330
1. Regenerate helm chart by running this script: https://sourcegraph.sgdev.org/github.com/sourcegraph/infrastructure@master/-/blob/kubernetes/generate.sh
1. (Verify your local git checkout has generated updates to the appropriate .yaml files)

```bash
gcloud container clusters get-credentials dogfood-cluster-7 --zone=us-central1-a
cd $PATH_TO_INFRASTUCTURE_REPO/on-prem/dogfood
helm upgrade sourcegraph ./sourcegraph-chart/
```

## Create staging environment from a branch

```bash
git push origin branch:staging/branch
```

Look in the #bots-staging channel for a notification when your staging environment is ready. (Your branch must first pass CI.)
