begin;
with last_campaign as (
select id from campaigns order by id desc limit 1
)
delete from events where campaign_id=(select id from last_campaign) or thread_id in (select ct.thread_id from campaigns_threads ct left join last_campaign on true where ct.campaign_id=last_campaign.id) and type in ('CreateThread', 'MergeThread', 'CloseThread', 'Review');

with last_campaign as (
     select id from campaigns order by id desc limit 1
),
updatec0 as (
     update campaigns set created_at=current_timestamp - interval '11 days', updated_at=current_timestamp - interval '11 days' where id=(select id from last_campaign)
),
updatec1 as (
update comments set created_at=current_timestamp - interval '11 days', updated_at=current_timestamp - interval '11 days' where campaign_id=(select id from last_campaign)
),
update0 as (
     update threads set created_at=current_timestamp - interval '11 days', updated_at=current_timestamp - interval '11 days' where id in (select thread_id from campaigns_threads ct left join last_campaign on true where campaign_id=last_campaign.id)
),
update1 as (
     update comments set created_at=current_timestamp - interval '11 days', updated_at=current_timestamp - interval '11 days' where thread_id in (select thread_id from campaigns_threads ct left join last_campaign on true where campaign_id=last_campaign.id)
),
update2 as (
     update events set created_at=current_timestamp - interval '11 days' where type='CreateThread' and thread_id in (select thread_id from campaigns_threads ct left join last_campaign on true where campaign_id=last_campaign.id)
),
all_threads as (
     select threads.id from threads left join last_campaign on true left join campaigns_threads ct on ct.campaign_id=last_campaign.id and ct.thread_id=threads.id
),
merged_threads as (select * from all_threads where id%2=0),
closed_threads as (select * from all_threads where id%5=3),
update_state0 as (update threads set state='MERGED' where id in (select id from merged_threads)),
update_state1 as (update threads set state='CLOSED' where id in (select id from closed_threads)),
insert_create_events as (
     insert into events(type,actor_user_id,created_at,thread_id)
     select 'CreateThread' as type,
     (select id from users where deleted_at is null order by (users.id+t.id)%31 asc limit 1),
     current_timestamp - interval '11 days', t.id
     from all_threads t
),
insert_merge_events as (
     insert into events(type,actor_user_id,created_at,thread_id)
     with _random as (select random() as n)
     select unnest('{MergeThread, CloseThread}'::text[]) as type,
            (select id from users where deleted_at is null order by (users.id+t.id)%31 asc limit 1),
            current_timestamp - (0.3+0.7*(select n from _random limit 1))*((1+(t.id % 7))::float/6)*(interval '10 days'), t.id
     from all_threads t where t.id%2=0
),
insert_close_events as (
     insert into events(type,actor_user_id,created_at,thread_id)
     select 'CloseThread' as type,
     (select id from users where deleted_at is null order by (users.id+t.id)%31 asc limit 1),
     current_timestamp - random()*((t.id % 7)::float/6)*(interval '10 days'), t.id
     from all_threads t where t.id%5=3
),
insert_review_events as (
     insert into events(type,actor_user_id,created_at,thread_id, data)
     select 'Review' as type,
     (select id from users where deleted_at is null order by (users.id+t.id)%31 asc limit 1),
     current_timestamp - random()*((t.id % 7)::float/6)*(interval '10 days'), t.id,
     (case when t.id % 3=0 then '{"state": "APPROVED"}' else '{"state": "CHANGES_REQUESTED"}' end)::jsonb
     from all_threads t where id%7<=4
),
scatter_diagnostic_events as (
     update events set created_at=current_timestamp - (0.6+0.4*random())*((id % 7)::float/6)*(interval '10 days')
     where type='AddDiagnosticToThread' and thread_id in (select thread_id from campaigns_threads ct left join last_campaign on true where campaign_id=last_campaign.id)
),
delete_most_diagnostic_events as (
     delete from events where id%9 != 1 and type='AddDiagnosticToThread' and thread_id in (select thread_id from campaigns_threads ct left join last_campaign on true where campaign_id=last_campaign.id)
)
select 1;
commit;
