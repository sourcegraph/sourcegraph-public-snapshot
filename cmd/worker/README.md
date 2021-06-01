# Worker

The worker service registers and performs background jobs that are not generally accessible via API. The worker service itself can be configured to run a specific task or set of tasks via environment variables. This allows site-administrators to scale background jobs individually depending on their unique instance needs.
