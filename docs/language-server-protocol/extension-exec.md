# `exec` LSP extension

The LSP `exec` extension allows the server to send a command (e.g. `git blame`) to the client for execution. The client responds with the result of executing the command (with the working directory set to the workspace root).

Open questions:

* Should the client state which commands are available (e.g., `git blame` but not `git commit`) in `ClientCapabilities`? Or just rely on the server noticing that the command returns an error and doing something else instead?

### Exec Request

_Request_:

* method: 'exec'
* params: `ExecParams` defined as follows:

```typescript
interface ExecParams {
  /**
   * The name of the program to run.
   */
  command: string;

  /**
   * The command-line arguments to the program.
   */
  arguments: string[];
}
```

_Response_:

* result: `ExecResult` defined as follows:

```typescript
interface ExecResult {
  /**
   * The stdout of the process.
   */
  stdout: string;

  /**
   * The stderr of the process.
   */
  stderr: string;

  /**
   * The exit code of the process.
   */
  exitCode: number;
}
```

* error: code and message set in case the command was not found or an exception occurs. It is not an error if the command is executed and exits with a nonzero exit code.
