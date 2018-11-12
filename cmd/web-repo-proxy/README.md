`web-repo-proxy` is a utility to serve short-lived Git repositories based on arbitrary code from a client. The initial application is to proxy the code from Stackoverflow posts through Git repositories so that browser extensions can show code intelligence for them. See [the original feature request](https://github.com/sourcegraph/sourcegraph/issues/423).

# Usage

`web-repo-proxy` accepts the following environment variables:

- `REPOSITORIES_ROOT`: the path on disk to store the repositories (Default: `"./repositories"`).
- `REPOSITORY_TTL`: the duration to save repositories before deleting them (Default: `"30m"`).
- `LISTENER_PORT`: the network port to listen on (Default: `4014`).
- `INSECURE_DEV`: for testing, runs the process as a local listener only, and enables the `/list-repositories` request (Default: `false`).

# Testing

To try it out locally:

```
cat >create-repository.json
{
  "repositoryName": "testRepo",
  "fileContents": {
    "asdf": "this is a file",
    "another_file": "files everywhere"
  }
}
^D

# Create a repository
curl -i http://localhost:4014/create-repository \
  -X POST \
  -H "Content-Type: application/json" \
  -d @create-repository.json

    HTTP/1.1 200 OK
    Date: Mon, 12 Nov 2018 20:01:38 GMT
    Content-Length: 66
    Content-Type: text/plain; charset=utf-8

    {"urlPath":"/repository/5b25c3bb/testRepo","goodUntil":1542054698}

# List all repositories (when run with INSECURE_DEV="true")
curl -i http://localhost:4014/list-repositories

HTTP/1.1 200 OK
Date: Mon, 12 Nov 2018 20:02:02 GMT
Content-Length: 54
Content-Type: text/plain; charset=utf-8

{"repository_paths":["/repository/5b25c3bb/testRepo"]}

# Clone the repository
git clone http://localhost:4014/repository/5b25c3bb/testRepo

    Cloning into 'testRepo'...

ls -l testRepo

    -rw-r--r--  1 fae  staff  16 Nov 12 15:02 another_file
    -rw-r--r--  1 fae  staff  14 Nov 12 15:02 asdf
```
