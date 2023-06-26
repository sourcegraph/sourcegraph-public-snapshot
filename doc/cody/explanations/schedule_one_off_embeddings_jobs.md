# Schedule one-off embeddings jobs

Adminstrators can schedule one-off embeddings jobs from the _Site Admin_ page. Open the _Site Admin_ page and select **Cody > Embeddings Jobs** from the left-hand navigation menu.

At the top of the page select the input field and search for the repository you want to index. Multiple repositories can be selected. Once you have selected the repositories, click on **Schedule Embedding**. The new jobs will be shown in the list of jobs below. The initial status of the jobs will be _queued_. 

<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/embeddings/schedule-one-off-jobs.png" class="screenshot" alt="schedule one-off embeddings jobs">

## FAQ

### I scheduled a one-off embeddings job but it is not showing up in the list of jobs. What happened?

There can be several reasons why a job is not showing up in the list of jobs:

- The repository is already queued or being processed
- A job for the same repository and the same revision already completed successfully
- Another job for the same repository has been queued for processing within the [embeddings.MinimumInterval](./code_graph_context.md#adjust-the-minimum-time-interval-between-automatically-scheduled-embeddings) time window 

### How do I stop a running embeddings job?

Jobs can be canceled while they are in state _queued_ or _processing_. To cancel a job, click on the _Cancel_ button of the job you want to cancel. The job will be marked for cancellation. Note that it might take a few seconds or minutes for the job to actually be canceled depending on the state of the job.
