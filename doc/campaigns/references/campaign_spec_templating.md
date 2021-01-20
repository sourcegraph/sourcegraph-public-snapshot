# Campaign spec templating

<style>
.markdown-body h2 { margin-top: 50px; }
.markdown-body pre.chroma { font-size: 0.75em; }
</style>

<aside class="experimental">
<span class="badge badge-experimental">Experimental</span> This feature is experimental and might change in the future. It's available in Sourcegraph 3.22 with <a href="https://github.com/sourcegraph/src-cli">Sourcegraph CLI</a> 3.21.5 and later.
</aside>

## Overview

[Certain fields](#fields-with-template-support) in a [campaign spec YAML](campaign_spec_yaml_reference.md) can include template variables to create even more powerful and performant campaigns.

Template variables in a campaign spec all have this form: `${{ <variable> }}`. They are evaluated before the execution of each entry in `steps` and allow accessing not only data from search results but also from previous steps.

Here is an example excerpt of a campaign spec that uses the template variables:

```yaml
on:
  - repositoriesMatchingQuery: lang:go fmt.Sprintf("%d", :[v]) patterntype:structural -file:vendor

steps:
  - run: comby -in-place 'fmt.Sprintf("%d", :[v])' 'strconv.Itoa(:[v])' ${{ join repository.search_result_paths " " }}
    container: comby/comby
  - run: goimports -w ${{ join previous_step.modified_files " " }}
    container: unibeautify/goimports
```

In this case, `${{ repository.search_result_paths }}` will be replaced with the relative-to-root-dir file paths of each search resulted yielded by `repositoriesMatchingQuery`. By using the [template helper function](#template-helper-functions) `join`, an argument list of whitespace-separated values is constructed. Before the step is executed the final `run` value would look close to this:

```yaml
run: comby -in-place 'fmt.Sprintf("%d", :[v])' 'strconv.Itoa(:[v])' cmd/src/main.go internal/fmt/fmt.go
```

The result is that `comby` only search and replaces in those files, instead of having to search through the complete repository.

The `${{ previous_step.modified_files }}` in the second step will be replaced by the list of files that the previous `comby` step modified. The final `run` value will look like this, if `comby` modified both of these files:

```yaml
run: goimports -w cmd/src/main.go internal/fmt/fmt.go
```

## Fields with template support

Template variables are supported in the following fields:

- [`steps.run`](campaign_spec_yaml_reference.md#steps-run)
- [`steps.env`](campaign_spec_yaml_reference.md#steps-run) values
- [`steps.files`](campaign_spec_yaml_reference.md#steps-run) values

## Template variables

The following template variables are available:

- `${{ repository.search_result_paths }}`

    Unique list of file paths relative to the repository root directory in which the search results of the `repositoriesMatchingQuery`s have been found.
- `${{ repository.name }}`

    Full name of the repository in which the step is being executed.
- `${{ previous_step.modified_files }}`

    List of files that have been modified by the previous step in `steps`. Empty if no files have been modified.
- `${{ previous_step.added_files }}`

    List of files that have been added by the previous step in `steps`. Empty if no files have been added.
- `${{ previous_step.deleted_files }}`

    List of files that have been deleted by the previous step in `steps`. Empty if no files have been deleted.
- `${{ previous_step.stdout }}`

    The complete output of the previous step on standard output.
- `${{ previous_step.stderr }}`

    The complete output of the previous step on standard error.

## Template helper functions

- `${{ join repository.search_result_paths "\n" }}`
- `${{ split repository.name "/" }}`

## Examples

Pass the exact list of search result file paths to a command:

```yaml
steps:
  - run: comby -in-place -config /tmp/go-sprintf.toml -f ${{ join repository.search_result_paths "," }}
    container: comby/comby
    files:
      /tmp/go-sprintf.toml: |
        [sprintf_to_strconv]
        match='fmt.Sprintf("%d", :[v])'
        rewrite='strconv.Itoa(:[v])'
```

Format and fix files after a previous step modified them:

```yaml
steps:
  - run: |
      find . -type f -name '*.go' -not -path "*/vendor/*" |\
      xargs sed -i 's/fmt.Println/log.Println/g'
    container: alpine:3
  - run: goimports -w ${{ join previous_step.modified_files " " }}
    container: unibeautify/goimports
```

Use the `steps.files` combined with template variables to construct files inside the container:

```yaml
steps:
  - run: |
      cat /tmp/search-results | while read file;
      do
        ruplacer --subvert whitelist allowlist --go ${file} || echo "nothing to replace";
        ruplacer --subvert blacklist denylist --go ${file} || echo "nothing to replace";
      done
    container: ruplacer
    files:
      /tmp/search-results: ${{ join repository.search_result_paths "\n" }}
```

Put information in environment variables, based on the output of previous step `steps.env` also 

```yaml
steps:
  - run: echo $LINTER_ERRROS >> linter_errors.txt
    container: alpine:3
    env:
      LINTER_ERRORS: ${{ previous_step.stdout }}
```

If you need to escape the `${{` and `}}` delimiters you can simply render them as string literals:


```yaml
steps:
  - run: cp /tmp/escaped.txt .
    container: alpine:3
    files:
      /tmp/escaped.txt: ${{ "${{" }} ${{ "}}" }}
```
