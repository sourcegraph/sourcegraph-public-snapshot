select
    commit_author_name,
    count(*) as commits
from search('
    type:commit
    repo:github.com/sourcegraph/sourcegraph$
    count:all
')
where not commit_author_name regexp '(R|r)enovate'
group by commit_author_name
order by commits desc
limit 10;
