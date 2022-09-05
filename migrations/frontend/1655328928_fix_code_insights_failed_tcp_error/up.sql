-- In 3.40 one transient error was accidentally flagged as non-retryable and would terminally fail. This simply
-- resets these jobs in the queue to try again after 3.41 ships.
update insights_query_runner_jobs set state = 'queued', failure_message = null, num_failures = 0
where state = 'failed' and failure_message like '%dial tcp%';
