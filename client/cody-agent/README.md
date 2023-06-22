# Cody agent

The `@sourcegraph/cody-agent` package implements a JSON-RPC server to interact
with Cody via stdout/stdin. This package is intended to be used by
non-ECMAScript clients such as the JetBrains and NeoVim plugins.

## Protocol

The protocol is defined in the file [`src/protocol.ts`](src/protocol.ts). The
TypeScript code is the single source of truth of what JSON-RPC methods are
supported in the protocol.

## Updating the protocol

Directly edit the TypeScript source code to add new JSON-RPC methods or add
properties to existing data structures.

The agent is a new project that is being actively worked on at the time of this
writing. The protocol is subject to breaking changes without notice. Please
let us know if you are implementing an agent client.

## Client bindings

There's no tool to automatically generate bindings for the Cody agent protocol.
Currently, clients have to manually write bindings for the JSON-RPC methods.

## Useful commands

- The command `pnpm run build-agent-binaries` builds standalone binaries for
  macOS, Linux, and Windows. By default, the binaries get written to the `dist/`
  directory. The destination directory can be configured with the environment
  variable `AGENT_EXECUTABLE_TARGET_DIRECTORY`.
- The command `pnpm run test` runs the agent against a minimized testing client.
  The tests are disabled in CI because they run against uses an actual Sourcegraph
  instance. Set the environment variables `SRC_ENDPOINT` and `SRC_ACCESS_TOKEN`
  to run the tests against an actual Sourcegraph instance.
  See the file [`src/index.test.ts`](src/index.test.ts) for a detailed but minimized example
  interaction between an agent client and agent server.

## Client implementations

- The Sourcegraph JetBrains plugin is defined in the sibling directory
  [`client/jetbrains`](../jetbrains/README.md). The file
  [`CodyAgentClient.java`](../jetbrains/src/main/java/com/sourcegraph/agent/CodyAgentClient.java)
  implements the client-side of the protocol.

## Miscellaneous notes

- By the nature of using JSON-RPC via stdin/stdout, both the agent server and
  client run on the same computer and there can only be one client per server.
  It's normal for both the client and server to be stateful processes. For
  example, the `connectionConfiguration/didChange` notification is sent from the
  client to the server to notify that subsequent requests should use the new
  connection configuration.
