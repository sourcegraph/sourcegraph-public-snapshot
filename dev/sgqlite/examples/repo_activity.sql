select
  strftime("%Y-%m", commit_committer_date) as month,
  count(*) as commits
from search('
    type:commit 
    repo:github.com/sourcegraph/sourcegraph 
    patterntype:literal 
    count:all
')
group by month
order by month desc
limit 12;
