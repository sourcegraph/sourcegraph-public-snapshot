# AWS CodeCommit integration with Sourcegraph

Sourcegraph integrates with [AWS CodeCommit](https://aws.amazon.com/codecommit/).

## AWS CodeCommit configuration

Sourcegraph supports syncing repositories from [AWS CodeCommit](https://aws.amazon.com/codecommit/). To add repositories from AWS CodeCommit:

1.  Go to the [site configuration editor](../admin/site_config/index.md).
2.  Press **Add AWS CodeCommit repositories**.
3.  Fill in the fields in the generated `awsCodeCommit` configuration option.

To see other optional AWS CodeCommit configuration settings, view [`awsCodeCommit` site config documentation](../admin/site_config/index.md#code-classlanguage-textawscodecommitconnection-object) or press Ctrl+Space or Cmd+Space in the site configuration editor.

## Browser extension

The [Sourcegraph browser extension](browser_extension.md) does not yet support AWS CodeCommit.
