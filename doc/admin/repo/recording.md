# Configuring command recording

Command recording allows site admins to view all git operations executed on a repository. 

When enabled, Sourcegraph will record metadata about all git commands run on a repository in Redis, including:

- Command executed (with sensitive information redacted)
- Execution time  
- Duration of execution
- Success state
- Output

This provides visibility into git operations performed by Sourcegraph on a repository, which can be useful for debugging and monitoring.

To enable command recording:

1. Go to Site Admin > Site Configuration
2. Add a `gitRecorder` object to the configuration object

```json
"gitRecorder": {
    // the amount of commands to record per repo  
    "size": 30,
  
    // repositories to record commands. This can either be a wildcard '*' 
    // to record commands for all repositories or a list of repositories
    "repos": ["*"],
  
    // git commands to exclude from recording. We exclude the 
    // commands below by default.
    "ignoredGitCommands": [
      "show",
      "rev-parse",
      "log",
      "diff",
      "ls-tree"
    ]
  }
```

Once enabled, site admins can view recorded commands for a repository via the repository's settings page in the Site Admin UI.

Recorded commands include information like start time, duration, exit status, command executed, directory, and output. Sensitive information like usernames, passwords, tokens are automatically redacted from the command and output.

Command recording provides visibility into Sourcegraph's interactions with repositories without requiring modifications to Sourcegraph's core Git operations.

### Potential Risks

Depending on the number of repositories and size of the recording set, enabling command recording could result in increased disk usage in Redis. 

Since recorded commands are stored in Redis, setting the `size` to a very large number or enabling recording on many repositories could cause the Redis database to fill up quickly.

When Redis is full, it may start evicting data which can impact other parts of Sourcegraph that rely on Redis. Sourcegraph may experience degraded performance or instability.

To avoid issues, we recommend proceeding with caution and starting with a smaller `size` and number of repositories first. Monitor your Redis memory usage over time and slowly increase the recording `size` and repositories. Tune the configuration based on your instance size and memory available.
