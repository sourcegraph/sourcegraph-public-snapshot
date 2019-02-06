import * as GQL from '../../../shared/src/graphql/schema'

export interface ExternalServiceMetadata {
    kind: GQL.ExternalServiceKind
    jsonSchemaId: string
    displayName: string
    defaultConfig: string
}

export const GITHUB_EXTERNAL_SERVICE: ExternalServiceMetadata = {
    kind: GQL.ExternalServiceKind.GITHUB,
    jsonSchemaId: 'site.schema.json#definitions/GitHubConnection',
    displayName: 'GitHub',
    defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/site_config/all#githubconnection-object

  "url": "https://github.com",

  // A token is required for access to private repos, but is also helpful for public repos
  // because it grants a higher hourly rate limit to Sourcegraph.
  // Create one with the repo scope at https://[your-github-instance]/settings/tokens/new
  "token": "",

  // Sync public repositories from https://github.com by adding them to "repos".
  // (This is not necessary for GitHub Enterprise instances)
  // "repos": [
  //     "sourcegraph/sourcegraph"
  // ]

}`,
}

export const ALL_EXTERNAL_SERVICES: ExternalServiceMetadata[] = [
    {
        kind: GQL.ExternalServiceKind.AWSCODECOMMIT,
        jsonSchemaId: 'site.schema.json#definitions/AWSCodeCommitConnection',
        displayName: 'AWS CodeCommit',
        defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/site_config/all#awscodecommitconnection-object

  "region": "",
  "accessKeyID": "",
  "secretAccessKey": ""
}`,
    },
    {
        kind: GQL.ExternalServiceKind.BITBUCKETSERVER,
        jsonSchemaId: 'site.schema.json#definitions/BitbucketServerConnection',
        displayName: 'Bitbucket Server',
        defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/site_config/all#bitbucketserverconnection-object

  "url": "https://bitbucket.example.com",

  // Create a personal access token with read scope at
  // https://[your-bitbucket-hostname]/plugins/servlet/access-tokens/add
  "token": ""
}`,
    },
    GITHUB_EXTERNAL_SERVICE,
    {
        kind: GQL.ExternalServiceKind.GITLAB,
        jsonSchemaId: 'site.schema.json#definitions/GitLabConnection',
        displayName: 'GitLab',
        defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/site_config/all#gitlabconnection-object

  "url": "https://gitlab.example.com",

  // Create a personal access token with api scope at
  // https://[your-gitlab-hostname]/profile/personal_access_tokens
  "token": ""
}`,
    },
    {
        kind: GQL.ExternalServiceKind.GITOLITE,
        jsonSchemaId: 'site.schema.json#definitions/GitoliteConnection',
        displayName: 'Gitolite',
        defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/site_config/all#gitoliteconnection-object

  "prefix": "gitolite.example.com/",
  "host": "git@gitolite.example.com"
}`,
    },
    {
        kind: GQL.ExternalServiceKind.PHABRICATOR,
        jsonSchemaId: 'site.schema.json#definitions/PhabricatorConnection',
        displayName: 'Phabricator',
        defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/site_config/all#phabricatorconnection-object

  "url": "https://phabricator.example.com",
  "token": "",
  "repos": []
}`,
    },
    {
        kind: GQL.ExternalServiceKind.OTHER,
        jsonSchemaId: 'site.schema.json#definitions/OtherExternalServiceConnection',
        displayName: 'Other',
        defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/site_config/all#otherexternalserviceconnection-object

  // Supported URL schemes are: http, https, git and ssh
  "url": "https://my-other-githost.example.com",

  // Repository clone paths may be relative to the url (preferred) or absolute.
  "repos": []
}`,
    },
]
