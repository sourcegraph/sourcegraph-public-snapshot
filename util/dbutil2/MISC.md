### Renaming a repository

Manual SQL. This assumes the repository in the `repo` table is under the new
name already and has been built under that name. This SQL only updates the build
data of *other* repositories that point to it. It also does not update
`symbol_stat`.

```sql
update ref set symbol_repo='github.com/docker/docker' where symbol_repo='github.com/dotcloud/docker';
update ref_author set symbol_repo='github.com/docker/docker' where symbol_repo='github.com/dotcloud/docker';
update ref_by_unit set symbol_repo='github.com/docker/docker' where symbol_repo='github.com/dotcloud/docker';
update refd_symbol_by_author_unit set symbol_repo='github.com/docker/docker' where symbol_repo='github.com/dotcloud/docker';
update unit_client_unit set symbol_repo='github.com/docker/docker' where symbol_repo='github.com/dotcloud/docker';

```
