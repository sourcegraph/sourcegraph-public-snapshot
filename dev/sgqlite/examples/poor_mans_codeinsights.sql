with first_commit_by_month as (
    select 
        month,
        commit_oid as first_commit
    from (
        select 
            strftime('%Y-%m', commit_committer_date) as month,
            commit_oid,
            row_number() over(
                partition by strftime('%Y-%m', commit_committer_date)
                order by commit_committer_date
            ) as rn
        from search('
            type:commit
            repo:github.com/sourcegraph/sourcegraph$
            count:all
        ')
    )
    where rn = 1
    order by month desc
)
select 
    month,
    (select sum(result_count) from search('
        type:file
        TODO
        count:all
        patterntype:literal
        repo:github.com/sourcegraph/sourcegraph@' || first_commit_by_month.first_commit
    )) as todos
from first_commit_by_month
limit 12;
