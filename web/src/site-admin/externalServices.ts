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
  "url": "https://github.com",

  // A token is required for access to private repos, but is also helpful for public repos
  // because it grants a higher hourly rate limit to Sourcegraph.
  "token": "<personal access token with repo scope (https://github.com/settings/tokens/new)>"
}`,
}

export const ALL_EXTERNAL_SERVICES: ExternalServiceMetadata[] = [
    {
        kind: GQL.ExternalServiceKind.AWSCODECOMMIT,
        displayName: 'AWS CodeCommit',
        defaultConfig: `{
  "region": "",
  "accessKeyID": "",
  "secretAccessKey": ""
}`,
    },
    {
        kind: GQL.ExternalServiceKind.BITBUCKETSERVER,
        displayName: 'Bitbucket Server',
        defaultConfig: `{
  "url": "https://bitbucket.example.com",
  "token": "<personal access token with read scope (https://[your-bitbucket-hostname]/plugins/servlet/access-tokens/add)>"
}`,
    },
    GITHUB_EXTERNAL_SERVICE,
    {
        kind: GQL.ExternalServiceKind.GITLAB,
        displayName: 'GitLab',
        defaultConfig: `{
  "url": "https://gitlab.example.com",
  "token": "<personal access token with api scope (https://[your-gitlab-hostname]/profile/personal_access_tokens)>"
}`,
    },
    {
        kind: GQL.ExternalServiceKind.GITOLITE,
        displayName: 'Gitolite',
        defaultConfig: `{
  "prefix": "gitolite.example.com/",
  "host": "git@gitolite.example.com"
}`,
    },
    {
        kind: GQL.ExternalServiceKind.PHABRICATOR,
        displayName: 'Phabricator',
        defaultConfig: `{
  "url": "https://phabricator.example.com",
  "token": "",
  "repos": []
}`,
    },
]
