This is an exploration into reimagining what the backfilling or potentially even all query running could look like for insights.

Below is a diagram to explain the flow this approach takes.

```mermaid
graph TD

subgraph worker
  Poll[Poll jobs] --> |New work found| RunBackfiller[Run Backfiller]
end

subgraph backfiller
  GenSearchJobs[Build Search Jobs] --> RunSearch[Search Runner]
  RunSearch --> SaveResults[Save Points]
end

RunBackfiller -->|Request backfill for series and repo| GenSearchJobs
SaveResults --> |Update results| RunBackfiller
```
