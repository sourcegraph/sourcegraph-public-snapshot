# P4.passwd setter tool

## How to use

1. Make sure you have the Sourcegraph CLI tool installed and configured – [instructions here](https://sourcegraph.com/github.com/sourcegraph/src-cli@main/-/blob/README.md)
2. On your Sourcegraph web UI, go to `user badge` (top-right) | `Site admin` | `Manage code hosts`, then look for your code host, click `Edit` next to it, and copy its ID from the page URL. It'll look like `RXg0YXJuZWxTZXJ3aWNlOjEzOA==`.
3. Save the script and make it executable.
4. Run [the script](p4passwd-setter.sh): `./p4passwd-setter.sh "CODE_HOST_ID" "NEW_P4_PASSWORD"`

The password will be updated for the code host and the script will give no output.

If something goes wrong, you'll get an error message.
