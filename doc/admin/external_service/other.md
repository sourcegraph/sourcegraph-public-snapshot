# Other Git repository hosts

Site admins can sync Git repositories on any Git repository host (by Git clone URL) with Sourcegraph so that users can search and navigate the repositories. Use this method only when your repository host is not named as a supported [external service](index.md).

To add Git repositories from any Git repository host:

1. Go to **User menu > Site admin**.
1. Open the **External services** page.
1. Press **+ Add external service**.
1. Press **Single Git repositories**.
1. Enter a **Display name** (such as the human-readable name of the repository host).
1. Set the `url` and `repos` fields in the JSON editor. Use Cmd/Ctrl+Space for completion, and [see configuration documentation below](#configuration).
1. Press **Add external service**.

## Constructing the `url` for SSH access

If your code host serves git repositories over SSH (e.g. Gerrit), make sure your Sourcegraph instance can connect to your code host over SSH:

```
docker exec $CONTAINER ssh -p $PORT $USER@$HOSTNAME
```

- $CONTAINER is the name or ID of your sourcegraph/server container
- $PORT is the port on which your code host's git server is listening for connections (Gerrit defaults to `29418`)
- $USER is your user on your code host (Gerrit defaults to `admin`)
- $HOSTNAME is the hostname of your code host from within the sourcegraph/server container (e.g. `gerrit.example.com`)

Here's an example for Gerrit:

```
docker exec sourcegraph ssh -p 29418 admin@gerrit.example.com
```

The `url` field is then

```json
  "url": "ssh://$USER@$HOSTNAME:$PORT"`
```

Here's an example for Gerrit:

```json
  "url": "ssh://admin@gerrit.example.com:29418",
```

## Adding repositories

For Gerrit, elements of the `repos` field are the same as the repository names. For example, a repository at https://gerrit.example.com/admin/repos/gorilla/mux will be `"gorilla/mux"` in the `repos` field.

Repositories must be listed individually:

```json
  "repos": [
    "gorilla/mux",
    "sourcegraph/sourcegraph"
  ]
```

## Experimental: src-expose

`src-expose` is a tool to periodically snapshot local directories and serve them as Git repositories over HTTP. This is a useful way to get code from other version control systems or textual artifacts from non version controlled systems (eg configuration) into Sourcegraph.

### Quick start

Start up a Sourcegraph instance

<pre class="pre-wrap start-sourcegraph-command"><code>docker run<span class="virtual-br"></span> --publish 7080:7080 --publish 127.0.0.1:3370:3370 --rm<span class="virtual-br"></span> --volume ~/.sourcegraph/config:/etc/sourcegraph<span class="virtual-br"></span> --volume ~/.sourcegraph/data:/var/opt/sourcegraph<span class="virtual-br"></span> sourcegraph/server:3.12.4</code></pre>

Pick a directory you want to export from, then run:

``` shell
wget https://storage.googleapis.com/sourcegraph-artifacts/src-expose/latest/darwin-amd64/src-expose
# For linux comment the above and uncomment the below
# wget https://storage.googleapis.com/sourcegraph-artifacts/src-expose/latest/linux-amd64/src-expose

chmod +x src-expose
./src-expose dir1 dir2 dir3
```

`src-expose` will output a configuration to use. It may scroll by quickly due to snapshot logging, so scroll up. However, this configuration should work:

``` javascript
 {
    // url is the http url to src-expose (listening on 127.0.0.1:3434)
    // url should be reachable by Sourcegraph.
    // "http://host.docker.internal:3434" works from Sourcegraph when using Docker for Desktop.
    "url": "http://host.docker.internal:3434",
    "repos": ["src-expose"]
}
```

Go to Admin > External services > Add external service > Single Git repositories. Input the above configuration. Your directories should now be syncing in Sourcegraph.

### Advanced configuration

The command line argument used by the quick start is for quickly validating the approach. However, you may have more complicated scenarios for snapshotting. In that case you can pass a YAML configuration file:

``` shell
src-expose -snapshot-config config.yaml
```

To see the configuration please consult `src-expose -help`. The [example.yaml](https://github.com/sourcegraph/sourcegraph/blob/master/dev/src-expose/example.yaml) also documents the possibilities.

### Serving git repositories

Alternatively you can serve git repositories. See `src-expose serve -help`.

## Configuration

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/external_service/other_external_service.schema.json">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/admin/external_service/other) to see rendered content.</div>
