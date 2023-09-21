# Configuring command recording

Command recording allows site admins to view all git operations executed on a repository. 

When enabled, Sourcegraph will record metadata about all git commands run on a repository, including:

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
