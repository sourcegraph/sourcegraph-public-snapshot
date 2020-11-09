# TODO

1. Find the right places in cloneRepo that should write to Postgres
2. Do the same for fetching.
3. Design the right GitserverStore method with nice semantics (i.e. upsert?)
4. Refactoring the RepositoryConnection GraphQL stuff to join against gitserver_repos instead of relying on repo.cloned
5. If that proves successful, remove repo.cloned and code that writes to it.

6. Document all other use cases we want to solve with gitserver leveraging Postgres
    - zoekt-index-server asking for all repos that have changed since the last time it asked
    
7. Separate idea to document and explore:
    - Instead of having fork and archived columns in repo, generalize this to something like a repo_tags table that can index other "tags" such as GitHub repo topics.

## Design

### Startup

On startup we need to handle rows which are marked as cloning but didn't get unmarked due to unclean shutdown.

``` sql
update gitserver_repos set state = 'not_cloned' where shard_id = 'gitserver-1' and state = cloning;
```

### Janitor

ensures that what is on disk matches what is in gitserver_repo.

one complication: if gitserver_repo.state == cloning and on disk state is not_cloned, then do not update state.
diff // select id from gitserver_repos where state = 'cloned';
update gitserver_repos set state = 'not_cloned' where shard_id = %s and state != 'cloning' and id = %s;

