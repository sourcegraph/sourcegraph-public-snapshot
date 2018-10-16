# Add repositories from any code host

To add repositories from any code host, use the `repos.list` site configuration setting. (In the Sourcegraph UI, you can go to the site configuration page and click **Add other repository** to fill in the fields in the generated `repos.list` configuration option.)

If cloning the repository requires authentication, make sure the [necessary credentials are set](#repositories-that-need-https-or-ssh-authentication).

Each item in `repos.list` has at least three parameters:

- `type`: Type of version control system. Currently `git` is only supported. Default is `git`.
- `url`: A valid git origin URL accessible from your Sourcegraph instance. (This URL may contain a username and password such as `https://user:password@example.com/my/repo`; they won't be displayed to non-admin users.)
- `path`: The path name of the repository, such as `my/repo`, `github.com/my/repo` (convention for GitHub.com repositories), `gitlab.example.com/my/repo` (convention for Gitlab EE/CE repositories, with `gitlab.example.com` being replaced by the actual host), etc.
- `links` (optional): Object containing URL formats for Sourcegraph to use when generating links back to the code host. `links` accepts four properties:
  - `repository`: URL specifying how to link to the repository, e.g., "https://example.com/myrepo".
  - `commit`: URL specifying how to link to commits. Use "{commit}" as a placeholder for a commit ID, e.g., "https://example.com/myrepo/view-commit/{commit}".
  - `tree`: URL specifying how to link to paths. Use "{path}" as a placeholder for the path, and "{rev}" as a placeholder for a revision, e.g., "https://example.com/myrepo@{rev}/browse/{path}".
  - `blob`: URL specifying how to link to files. Use "{path}" as a placeholder for the path, and "{rev}" as a placeholder for a revision, e.g., "https://example.com/myrepo@{rev}/blob/{path}".

When your server starts up, it will go through the repositories listed in `repos.list` and automatically clone and make them available on your Sourcegraph.

---

## Repositories that need HTTP(S) or SSH authentication

If authentication is required to `git clone` the repository clone URLs that `repos.list` specifies, then you must provide the credentials to the container.

### SSH authentication (config, keys, `known_hosts`)

The container consults its own file system (in the standard locations) for SSH configuration, private keys, and `known_hosts`. Upon container start, it copies all files from `/etc/sourcegraph/ssh` into its own `$HOME/.ssh` directory.

To provide SSH authentication configuration to the container, assuming you're using the default `--volume $HOME/.sourcegraph/config:/etc/sourcegraph`, follow these steps:

1.  Create files at `$HOME/.sourcegraph/config/ssh/config`, `$HOME/.sourcegraph/config/ssh/known_hosts`, etc., on the host machine as desired to configure SSH.
1.  Start (or restart) the container.

To configure the container to use the same SSH as your user account on the host machine, you can also run `cp -R $HOME/.ssh $HOME/.sourcegraph/config/ssh`.

### HTTP(S) authentication via netrc

The easiest way to specify HTTP(S) authentication for repositories is to include the username and password in the clone URL itself (in the `repos.list` entry's `url` property), such as `https://user:password@example.com/my/repo`. These credentials won't be displayed to non-admin users.

Otherwise, the container consults the `$HOME/.netrc` files on its own file system for HTTP(S) authentication. The `.netrc` file is a standard way to specify authentication used to connect to external hosts.

To provide HTTP(S) authentication, assuming you're using the default `--volume $HOME/.sourcegraph/config:/etc/sourcegraph`, follow these steps:

1.  Create a file at `$HOME/.sourcegraph/config/netrc` on the host machine that contains lines of the form `machine example.com login alice password mypassword` (replacing `example.com`, `alice`, and `mypassword` with the actual values).
1.  Start (or restart) the container.
