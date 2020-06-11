# Campaign templates

> TODO(sqs)

A campaign template describes how to produce the campaign's changesets. The format resembles that of [GitHub Actions workflows](https://help.github.com/en/actions/reference/workflow-syntax-for-github-actions).

``` yaml
# What repositories do you want to change? The script is executed independently in each root specified here. A root is usually a repository on a particular branch, but it can also be a subdirectory of a repository (e.g., for monorepos).
roots:
  - query: repo:github.com/sourcegraph/ file:package.json react-router-dom
  - repository: github.com/acmecorp/foo
    branch: dev
    path: pkg/webapp

# The Docker container in which steps (that don't specify a container) will be executed.
container: node:14

# The name of the branch where the changes are pushed.
branch: upgrade-react-router-dom

# The steps are executed sequentially in each root.
steps:
  - name: Upgrade react-router-dom
    run: yarn upgrade -L react-router-dom
  - name: Rewrite <Redirect> to <Navigate>
    run: comby '<Redirect :[1] />' '<Navigate :[1] />' .tsx -i
    container: comby/comby
```



### Campaign spec

You never need to read or edit a campaign spec. Instead, you'll create a [campaign template](#campaign-templates), then run `src campaign apply -template foo.campaign.yml -campaign <ID> [-preview]` to generate, upload, and apply the campaign spec. In case you're curious, though, here's what a campaign spec looks like:

```json
{
  "changesets": [
    {
      "base": {
        "repository": "<ID>",
        "branch": "master"
      },
      "head": {
        "repository": "<ID>",
        "branch": "upgrade-react-router-dom",
        "commit": "35da342b5fb78d315713efebb3431095c1d1de16"
      },
      "title": "Upgrade react-router-dom, migrate <Redirect> -> <Navigate>",
      "description": "...",
      "commits": [
        {
          "message": "{{.Title}}"",
          "patch": "..." // unified diff format
        }
      ]
    },
    { ... }, // other changesets
  ]
}
```
