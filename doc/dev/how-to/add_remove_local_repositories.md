# How to add or remove repositories from your local environment

Depending on your needs, you may want to add or remove repositories from your local environment.  
Fewer repositories will result in a quicker start time and less resources used on your local computer.

The default repositories to load are located in a file named `external-services-config.json`.  
This file is located here: `/dev-private/enterprise/dev/external-services-config.json` on your local development environment.

Your local development environment must be restarted to pick up any changes you have made.

Find this section of the code: 

```
  "GITHUB": [
    {
      "authorization": {},
      "url": "https://github.com",
      "token": "f189fe46ca1901b4988936cbc671f8305715ed73", // unprivileged sqs-test user
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
    {
      "url": "https://ghe.sgdev.org",
      "token": "1ec8973c1debb3232468b514b1a850dba87ee29a",
      "repositoryQuery": ["org:sourcegraph gorilla"]
    }
  ],
```

**BEFORE** you make any changes, it is **strongly recommended** that you make a local backup of this file so that you can easily revert any changes if needed.

To add or remove repositories, change the entires in the **repos** array:

```
      "repos": [
        "sourcegraph/sourcegraph",
        "hashicorp/go-multierror",
        "hashicorp/errwrap",
        "sourcegraph-testing/etcd",
        "sourcegraph-testing/tidb",
        "sourcegraph-testing/titan",
        "sourcegraph-testing/zap"
      ],
```

**NOTE:** The last item in the list does **NOT** have a comma at the end.

## Adding a repository

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

## Removing a repository

To remove one or more repositories from your local environment, simply remove those lines.  
Remember to ensure a comma is present after every item in the list except the last item.

For example, to remove the "sourcegraph-testing/zap" and "sourcegraph-testing/etcd" repositories:

```
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

