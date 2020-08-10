# Example: Using ESLint to automatically migrate to a new TypeScript version

> NOTE: This documentation describes the current work-in-progress version of campaigns. [Click here](https://docs.sourcegraph.com/@3.18/user/campaigns) to read the documentation for campaigns in Sourcegraph 3.18.

<!-- TODO(sqs): update for new campaigns flow -->

Our goal for this campaign is to convert all TypeScript code synced to our Sourcegraph instance to make use of new TypeScript features. To do this we convert the code, then update the TypeScript version.

To convert the code we install and run ESLint with the desired `typescript-eslint` rules, using the [`--fix` flag](https://eslint.org/docs/user-guide/command-line-interface#fix) to automatically fix problems. We then update the TypeScript version using [`yarn upgrade`](https://legacy.yarnpkg.com/en/docs/cli/upgrade/).

The first thing we need is a Docker container in which we can freely install and run ESLint. Here is the `Dockerfile`:

```Dockerfile
FROM node:12-alpine3.10
CMD package_json_bkup=$(mktemp) && \
  cp package.json $package_json_bkup && \
  yarn -s --non-interactive --pure-lockfile --ignore-optional --ignore-scripts --ignore-engines --ignore-platform --no-progress && \
  yarn add --ignore-workspace-root-test --non-interactive --ignore-optional --ignore-scripts --ignore-engines --ignore-platform --pure-lockfile -D @typescript-eslint/parser @typescript-eslint/eslint-plugin --no-progress eslint && \
  node_modules/.bin/eslint \
  --fix \
  --plugin @typescript-eslint \
  --parser @typescript-eslint/parser \
  --parser-options '{"ecmaVersion": 8, "sourceType": "module", "project": "tsconfig.json"}' \
  --rule '@typescript-eslint/prefer-optional-chain: 2' \
  --rule '@typescript-eslint/no-unnecessary-type-assertion: 2' \
  --rule '@typescript-eslint/no-unnecessary-type-arguments: 2' \
  --rule '@typescript-eslint/no-unnecessary-condition: 2' \
  --rule '@typescript-eslint/no-unnecessary-type-arguments: 2' \
  --rule '@typescript-eslint/prefer-includes: 2' \
  --rule '@typescript-eslint/prefer-readonly: 2' \
  --rule '@typescript-eslint/prefer-string-starts-ends-with: 2' \
  --rule '@typescript-eslint/prefer-nullish-coalescing: 2' \
  --rule '@typescript-eslint/no-non-null-assertion: 2' \
  '**/*.ts'; \
  mv $package_json_bkup package.json; \
  rm -rf node_modules; \
  yarn upgrade --latest --ignore-workspace-root-test --non-interactive --ignore-optional --ignore-scripts --ignore-engines --ignore-platform --no-progress typescript && \
  rm -rf node_modules
```

When turned into an image and run as a container, the instructions in this Dockerfile will do the following:

1. Copy the current `package.json` to a backup location so that we can install ESLint without changes to the original `package.json`
1. Install all dependencies & add `eslint` with the `typescript-eslint` plugin
1. Run `eslint --fix` with a set of TypeScript rules to detect and fix problems over all `*.ts` files
1. Restore the original `package.json` from its backup location
1. Run `yarn upgrade` to update the `typescript` version

Before we can run it as an action we need to turn it into a Docker image, by running the following command in the directory where the `Dockerfile` was saved:

```sh
docker build -t eslint-fix-action .
```

That builds a Docker image and names it `eslint-fix-action`.

Once that is done we're ready to define our action:

```json
{
  "scopeQuery": "repohasfile:yarn\\.lock repohasfile:tsconfig\\.json",
  "steps": [
    {
      "type": "docker",
      "image": "eslint-fix-action"
    }
  ]
}
```

The `"scopeQuery"` ensures that the action will only be run over repositories containing both a `yarn.lock` and a `tsconfig.json` file. This narrows the scope down to only the TypeScript projects in which we can use `yarn` to install dependencies. Feel free to narrow it down further by using more filters, such as `repo:my-project-name` to only run over repositories that have `my-project-name` in their name.

The action only has a single step to execute in each repository: it runs the Docker container we just built (called `eslint-fix-action`) on the machine on which `src` is executed.

Save that definition in a file called `eslint-fix-typescript.action.json` and we're ready to execute it.

First we make sure that we match all the repositories we want:

```sh
src actions scope-query -f eslint-fix-typescript.action.json
```

If that list looks good, we're ready to execute the action:

```sh
src actions exec -f eslint-fix-typescript.action.json | src campaign patchset create-from-patches
```

You should now see that the Docker container we built is being executed in a local, temporary copy of each repository.

After executing, the patches it produced will be turned into a patch set you can preview on our Sourcegraph instance.

## Caching dependencies across multiple steps using `cacheDirs`

If you find yourself writing an action definition that relies on a project's dependencies to be installed for every step it can be helpful to cache these dependencies in a directory outside of the repository.

For `"docker"` steps you can use the `cacheDirs` attribute:

```json
{
  "scopeQuery": "repohasfile:package.json",
  "steps": [
    {
      "type": "docker",
      "image": "yarn-install-dependencies",
      "cacheDirs": [ "/cache" ]
    },
    {
      "type": "docker",
      "image": "eslint-fix-action",
      "cacheDirs": [ "/cache" ]
    },
    {
      "type": "docker",
      "image": "upgrade-typescript",
      "cacheDirs": [ "/cache" ]
    }
  ]
}
```

This defines three separate steps that all use the same `"cacheDirs"`. The specified directory `"/cache"` will be created on the machine on which `src` is executing in a temporary location and mounted in each of the three containers under `/cache`.

As an example, here is the first part of `Dockerfile` that makes use of this `/cache` directory when installing dependencies with `yarn`:

```
FROM node:12-alpine3.10
VOLUME /cache
ENV YARN_CACHE_FOLDER=/cache/yarn
ENV NPM_CONFIG_CACHE=/cache/npm
CMD yarn -s --non-interactive --pure-lockfile --ignore-optional --ignore-scripts --ignore-engines --ignore-platform --no-progress

# [...]
```

This uses `/cache` as the `YARN_CACHE_FOLDER` and `NPM_CONFIG_CACHE` folder. It installs the dependencies with `yarn`, thus populating the `/cache` folder.

Subsequent action steps that use this preamble in their `Dockerfile` will run faster because they can leverage the cache folder.
