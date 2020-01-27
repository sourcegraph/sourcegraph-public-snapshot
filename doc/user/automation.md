# Automation

> Automation is currently available in private beta for select enterprise customers.

[Sourcegraph automation](https://about.sourcegraph.com/product/automation) allows large-scale code changes across many repositories and different code hosts.

## Configuration

In order to use the Automation preview, a site-admin of your Sourcegraph instance must enable it in the site configuration settings e.g. `sourcegraph.example.com/site-admin/configuration`

```json
{
  "experimentalFeatures": {
      "automation": "enabled"
  }
}
```

## Usage

There are two types of Automation campaigns:

- Manual campaigns to which you can manually add changesets (pull requests) and track their progress
- Campaigns created from a set of patches

### Creating a manual campaign

1. Go to `/campaigns` on your Sourcegraph instance and click on the "New campaign" button
2. Fill in a name for the campaign and a description
3. Create the campaign
4. Track changesets by adding them to the campaign through the form on the Campaign page

### Creating a campaign from a patches

**Required**: The [`src` CLI tool](https://github.com/sourcegraph/src-cli). 

The first thing we need is a definition of an "action". An action is what produces a patch and describes what commands and Docker containers to run over which repositories.

Example:

```json
{
  "scopeQuery": "repo:go-* -repohasfile:INSTALL.md",
  "steps": [
    {
      "type": "command",
      "args": ["sh", "-c", "echo '# Installation' > INSTALL.md"]
    },
    {
      "type": "docker",
      "dockerfile": "FROM alpine:3 \n CMD find /work -iname '*.md' -type f | xargs -n 1 sed -i s/this/that/g"
    },
    {
      "type": "docker",
      "image": "golang:1.13-alpine",
      "args": ["go", "fix", "/work/..."]
    }
  ]
}
```

This action runs over every repository that has `go-` in its name and doesn't have an `INSTALL.md` file.

The first step creates an `INSTALL.md` file by running `sh` in each repository on the machine on which `src` is executed.

The second step builds a Docker image from the specified `"dockerfile"` and starts a container with this image in which the repository is mounted under `/work`.

The third step pulls the `golang:1.13-alpine` image from Docker hub, starts a container from it and runs `go fix /work/...` in it.

---

If you are looking to run automation on a larger scale in the local dev environment, follow the [guide on automation development](../dev/automation_development.md).
