select
    commit_author_name,
    count(*) as commits
from search('
    type:commit
    repo:github.com/sourcegraph/sourcegraph$
    count:all
')
where date(commit_committer_date) > date('now', '-1 year')
group by commit_author_name
order by commits desc
limit 10;
