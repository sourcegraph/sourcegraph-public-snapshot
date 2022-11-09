# P4.passwd setter tool

## How to use

1. Install the Sourcegraph CLI tool if you don't have it – [CLI quickstart](https://docs.sourcegraph.com/cli/quickstart)
2. Set your `SRC_ENDPOINT` environment variable to your Sourcegraph instance URL
3. On your Sourcegraph web UI, go to `user badge` (top-right) | `Settings` | `Access tokens` and generate a token.
4. Set your `SRC_ACCESS_TOKEN` environment variable to your access token
5. On your Sourcegraph web UI, go to `user badge` (top-right) | `Site admin` | `Manage code hosts`, then look for your code host, click `Edit` next to it, and copy its ID from the page URL. It'll look like `RXg0YXJuZWxTZXJ3aWNlOjEzOA==`.
6. Save the script and make it executable.
7. Run [the script](p4passwd-setter.sh): `./p4passwd-setter.sh "CODE_HOST_ID" "NEW_P4_PASSWORD"`

The password will be updated for the code host and the script will give no output.

If something goes wrong, you'll get an error message.
