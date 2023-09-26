# Configuring command recording

Command recording allows site admins to view all git operations executed on a repository. 

![Recorded commands page](https://sourcegraphstatic.com/docs/images/admin/config/command_recording.png)

When enabled, Sourcegraph will record metadata about all git commands run on a repository in Redis, including:

- Command executed (with sensitive information redacted)
- Execution time  
- Duration of execution
- Success state
- Output

This provides visibility into git operations performed by Sourcegraph on a repository, which can be useful for debugging and monitoring.

To enable command recording:

- Go to [Site Admin > Site Configuration](../config/site_config.md)
- Add a `gitRecorder` object to the configuration object

```jsonc
{
  // [...]

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
}
```

Once enabled, site admins can view recorded commands for a repository via the repository's settings page in the Site Admin UI.

Recorded commands include information like start time, duration, exit status, command executed, directory, and output. Sensitive information like usernames, passwords, and tokens are automatically redacted from the command and output.

Command recording provides visibility into Sourcegraph's interactions with repositories without requiring modifications to Sourcegraph's core Git operations.

### Potential risks

Enabling command recording will increase disk usage in Redis, potentially in a significant manner, depending on the number of repositories and the size of the recording set.

Since recorded commands are stored in Redis, setting the `size` to a very large number or enabling recording on many repositories could cause the Redis database to fill up quickly.

Depending on your configuration, Redis might evict data from the database when it is full, impacting other parts of Sourcegraph that rely on Redis. This could cause Sourcegraph to experience degraded performance or instability.

To avoid issues, we recommend proceeding with caution and starting with a smaller `size` and number of repositories first. Monitor your Redis memory usage over time and slowly increase the recording `size` and repositories. Tune the configuration based on your instance size and memory available.
