import * as GQL from '../../../shared/src/graphql/schema'

export const ALL_EXTERNAL_SERVICES: { kind: GQL.ExternalServiceKind; displayName: string }[] = [
    { kind: GQL.ExternalServiceKind.AWSCODECOMMIT, displayName: 'AWS CodeCommit' },
    { kind: GQL.ExternalServiceKind.BITBUCKETSERVER, displayName: 'Bitbucket Server' },
    { kind: GQL.ExternalServiceKind.GITHUB, displayName: 'GitHub' },
    { kind: GQL.ExternalServiceKind.GITLAB, displayName: 'GitLab' },
    { kind: GQL.ExternalServiceKind.GITOLITE, displayName: 'Gitolite' },
    { kind: GQL.ExternalServiceKind.PHABRICATOR, displayName: 'Phabricator' },
]
