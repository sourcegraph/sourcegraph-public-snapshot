# How to add or remove repositories from your local environment

Depending on your needs, you may want to add or remove repositories from your local environment.  
Fewer repositories will result in a quicker start time and less resources used on your local computer.

The default repositories to load are located in a file named `external-services-config.json`.  
This file is located here: `/dev-private/enterprise/dev/external-services-config.json` in your clone of the 'dev-private' repository.

Your local development environment must be restarted to pick up any changes you have made.

There are two (2) main areas to that control the behavior of including repositories in your instance:
- The **repos** array
- The **repositoryQuery** value

## Changing the repos array

Find this section of the code: 

```
  "GITHUB": [
    {
      "authorization": {},
      "url": "https://github.com",
      "token": xxxxxx
      "repositoryQuery": ["affiliated"],
      "repos": [
        "sourcegraph/sourcegraph",
        "hashicorp/go-multierror",
        "hashicorp/errwrap",
        "sourcegraph-testing/etcd",
        "sourcegraph-testing/tidb",
        "sourcegraph-testing/titan",
        "sourcegraph-testing/zap"
      ],
      "cloudDefault": true
    },
      ... CODE REMOVED FOR BREVITY ...
  ],
```

**BEFORE** you make any changes, it is **strongly recommended** that you make a local backup of this file so that you can easily revert any changes if needed.

To add or remove repositories, change the entries in the **repos** array, please note the last item in the list does **NOT** have a comma at the end.

### Adding a repository

For example, to add a repository named `sourcegraph-test/my-new-feature` to your local environment, you would add a line into the array as shown below.  
Remember to add a comma to the end of the previous line:

```
      "repos": [
        "sourcegraph/sourcegraph",
        "hashicorp/go-multierror",
        "hashicorp/errwrap",
        "sourcegraph-testing/etcd",
        "sourcegraph-testing/tidb",
        "sourcegraph-testing/titan",
        "sourcegraph-testing/zap",
        "sourcegraph-test/my-new-feature"
      ],
```

After you have made your edits, remember to save your changes.

### Removing a repository

To remove one or more repositories from your local environment, simply remove those lines.  
Remember to ensure a comma is present after every item in the list except the last item.

For example, to remove the "sourcegraph-testing/zap" and "sourcegraph-testing/etcd" repositories:

```
      "repos": [
        "sourcegraph/sourcegraph",
        "hashicorp/go-multierror",
        "hashicorp/errwrap",
        "sourcegraph-testing/tidb",
        "sourcegraph-testing/titan",
      ],
```

After you have made your edits, remember to save your changes.

## Changing the repositoryQuery value

The possible values for are documented here:
https://docs.sourcegraph.com/admin/external_service/github#configuration

Please be aware that the default value is 'affiliated'.  
This value means that all repositories affiliated with the configured token's user are mirrored.
- Private repositories with read access
- Public repositories owned by the user or their orgs
- Public repositories with write access

For instance, to change this value so that no repositories are included, we use 'none':

"repositoryQuery": ["none"],

After you have made your edits, remember to save your changes.
