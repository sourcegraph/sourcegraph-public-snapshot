import * as GQL from '../../../shared/src/graphql/schema'

export interface ExternalServiceMetadata {
    kind: GQL.ExternalServiceKind
    displayName: string
    defaultConfig: string
}

export const GITHUB_EXTERNAL_SERVICE: ExternalServiceMetadata = {
    kind: GQL.ExternalServiceKind.GITHUB,
    displayName: 'GitHub',
    defaultConfig: `{
  // Hover over fields for documentation and Ctrl+Space for auto-completion.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/site_config/all#githubconnection-object

  "url": "https://github.com",

  // A token is required for access to private repos, but is also helpful for public repos
  // because it grants a higher hourly rate limit to Sourcegraph.
  "token": "<personal access token with repo scope (https://github.com/settings/tokens/new)>",

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
        displayName: 'AWS CodeCommit',
        defaultConfig: `{
  // Hover over fields for documentation and Ctrl+Space for auto-completion.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/site_config/all#awscodecommitconnection-object

  "region": "",
  "accessKeyID": "",
  "secretAccessKey": ""
}`,
    },
    {
        kind: GQL.ExternalServiceKind.BITBUCKETSERVER,
        displayName: 'Bitbucket Server',
        defaultConfig: `{
  // Hover over fields for documentation and Ctrl+Space for auto-completion.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/site_config/all#bitbucketserverconnection-object

  "url": "https://bitbucket.example.com",
  "token": "<personal access token with read scope (https://[your-bitbucket-hostname]/plugins/servlet/access-tokens/add)>"
}`,
    },
    GITHUB_EXTERNAL_SERVICE,
    {
        kind: GQL.ExternalServiceKind.GITLAB,
        displayName: 'GitLab',
        defaultConfig: `{
  // Hover over fields for documentation and Ctrl+Space for auto-completion.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/site_config/all#gitlabconnection-object

  "url": "https://gitlab.example.com",
  "token": "<personal access token with api scope (https://[your-gitlab-hostname]/profile/personal_access_tokens)>"
}`,
    },
    {
        kind: GQL.ExternalServiceKind.GITOLITE,
        displayName: 'Gitolite',
        defaultConfig: `{
  // Hover over fields for documentation and Ctrl+Space for auto-completion.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/site_config/all#gitoliteconnection-object

  "prefix": "gitolite.example.com/",
  "host": "git@gitolite.example.com"
}`,
    },
    {
        kind: GQL.ExternalServiceKind.PHABRICATOR,
        displayName: 'Phabricator',
        defaultConfig: `{
  // Hover over fields for documentation and Ctrl+Space for auto-completion.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/site_config/all#phabricatorconnection-object

  "url": "https://phabricator.example.com",
  "token": "",
  "repos": []
}`,
    },
]
