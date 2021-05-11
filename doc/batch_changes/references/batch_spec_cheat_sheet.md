# Batch spec cheat sheet

<style>
.markdown-body h2 { margin-top: 50px; }
.markdown-body pre.chroma { font-size: 0.75em; }
</style>

## Overview

There are some common patterns that we reuse all the time when writing [batch specs](batch_spec_yaml_reference.md). In this document we collect these pattern to make it easy for others to copy and reuse them.

Since most of the examples here make use of [batch spec templating](batch_spec_templating.md) make sure to also take a look at that page.

### Loop over search result paths in shell script

```yaml
on:
  - repositoriesMatchingQuery: OLD-VALUE

steps:
  - run: |
      for file in "${{ join repository.search_result_paths " " }}";
      do
        sed -i 's/OLD-VALUE/NEW-VALUE/g;' "${file}"
      done
    container: alpine:3
```

### Put search result paths in file and loop over them

```yaml
on:
  - repositoriesMatchingQuery: OLD-VALUE

steps:
  - run: |
      cat /tmp/search-results | while read file;
      do
        sed -i 's/OLD-VALUE/NEW-VALUE/g;' "${file}"
      done
    container: alpine:3
    files:
      /tmp/search-results: ${{ join repository.search_result_paths "\n" }}
```

### Use search result paths as arguments for single command

```yaml
on:
  - repositoriesMatchingQuery: lang:go fmt.Sprintf("%d", :[v]) patterntype:structural -file:vendor count:10

steps:
  - run: comby -in-place 'fmt.Sprintf("%d", :[v])' 'strconv.Itoa(:[v])' ${{ join repository.search_result_paths " " }}
    container: comby/comby
```

### Format files modified by previous step

```yaml
steps:
  - run: comby -in-place 'fmt.Sprintf("%d", :[v])' 'strconv.Itoa(:[v])' ${{ join repository.search_result_paths " " }}
    container: comby/comby
  - run: goimports -w ${{ join previous_step.modified_files " " }}
    container: unibeautify/goimports
```

### Dynamically set branch name based on workspace

```yaml
workspaces:
  - rootAtLocationOf: package.json
    in: github.com/sourcegraph/*

steps:
  # [... other steps ... ]
  - run:  if [[ -f "package.json" ]]; then cat package.json | jq -j .name; fi
    container: jiapantw/jq-alpine:latest
    outputs:
      projectName:
        value: ${{ step.stdout }}

changesetTemplate:
  # [...]

  # If we have an `outputs.projectName` we use it, otherwise we append the path
  # of the workspace. If the path is emtpy (as is the case in the root folder),
  # we ignore it.
  branch: |
    ${{ if eq outputs.projectName "" }}
    ${{ join_if "-" "thorsten/workspace-discovery" (replace steps.path "/" "-") }}
    ${{ else }}
    thorsten/workspace-discovery-${{ outputs.projectName }}
    ${{ end }}
```

### Process search result paths with script

```yaml
steps:
  - run: |
      for result in "${{ join repository.search_result_paths " " }}"; do
        ruby /tmp/script "${result}" > "${result}.new"
        mv ${result}.new "${result}"
      done;
    container: ruby
    files:
      /tmp/script: |
        #! /usr/bin/env ruby
        require 'yaml';
        content = YAML.load(ARGF.read)
        content['batchchanges'] = 'say hello'
        puts YAML.dump(content)
```

### Use separate file as config file for command

```yaml
steps:
  - run: comby -in-place -matcher .go -config /tmp/comby-conf.toml -f ${{ join repository.search_result_paths "," }}
    container: comby/comby
    files:
      /tmp/comby-conf.toml: |
        [log_to_log15]
        match='log.Printf(":[format]", :[args])'
        rewrite='log15.Warn(":[format]", :[args])'
        rule='where
        rewrite :[format] { "%:[[_]] " -> "" },
        rewrite :[format] { " %:[[_]]" -> "" },
        rewrite :[args] { ":[arg~[a-zA-Z0-9.()]+]" -> "\":[arg]\", :[arg]" }'
```

### Publish only changesets on specific branches

```yaml
changesetTemplate:
  # [...]
  published:
    - github.com/my-org/my-repo@my-branch-name: draft
```

### Create new files in repository

```yaml
steps:
  - run: cat /tmp/global-gitignore >> .gitignore
    container: alpine:3
    files:
      /tmp/global-gitignore: |
        # Vim
        *.swp
        # JetBrains/IntelliJ
        .idea
        # Emacs
        *~
        \#*\#
        /.emacs.desktop
        /.emacs.desktop.lock
        .\#*
        .dir-locals.el
```

### Execute steps only in repositories matching name

```yaml
steps:
  # [... other steps ...]
  - run: echo "name contains sourcegraph-testing" >> message.txt
    if: ${{ matches repository.name "*sourcegraph-testing*" }}
    container: alpine:3
```

### Execute steps based on output of previous command

```yaml
steps:
  - run:  if [[ -f "go.mod" ]]; then echo "true"; else echo "false"; fi
    container: alpine:3
    outputs:
      goModExists:
        value: ${{ step.stdout }}

  - run: go fmt ./...
    container: golang
    if: ${{ outputs.goModExists }}
```
