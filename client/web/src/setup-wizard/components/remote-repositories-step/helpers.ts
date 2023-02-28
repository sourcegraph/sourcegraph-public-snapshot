import { mdiBitbucket, mdiGithub, mdiGitlab, mdiAws, mdiGit } from '@mdi/js'

import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'

import { GetCodeHostsResult } from '../../../graphql-operations'

export const getCodeHostIcon = (codeHostType: ExternalServiceKind | null): string => {
    switch (codeHostType) {
        case ExternalServiceKind.GITHUB:
            return mdiGithub
        case ExternalServiceKind.BITBUCKETCLOUD:
            return mdiBitbucket
        case ExternalServiceKind.BITBUCKETSERVER:
            return mdiBitbucket
        case ExternalServiceKind.GITLAB:
            return mdiGitlab
        case ExternalServiceKind.GITOLITE:
            return mdiGit
        case ExternalServiceKind.AWSCODECOMMIT:
            return mdiAws
        case ExternalServiceKind.AZUREDEVOPS:
            return mdiGit
        default:
            // TODO: Add support for other code host
            return ''
    }
}

export const getCodeHostName = (codeHostType: ExternalServiceKind | null): string => {
    switch (codeHostType) {
        case ExternalServiceKind.GITHUB:
            return 'GitHub'
        case ExternalServiceKind.GITLAB:
            return 'GitLab'
        case ExternalServiceKind.BITBUCKETCLOUD:
            return 'BitBucket.org'
        case ExternalServiceKind.BITBUCKETSERVER:
            return 'BitBucket Server'
        case ExternalServiceKind.AWSCODECOMMIT:
            return 'AWS Code Commit'
        case ExternalServiceKind.GITOLITE:
            return 'Gitolite'
        case ExternalServiceKind.GERRIT:
            return 'Gerrit'

        default:
            // TODO: Add support for other code host
            return 'Unknown'
    }
}

export const getCodeHostURLParam = (codeHostType: ExternalServiceKind): string => codeHostType.toString().toLowerCase()

export const getCodeHostKindFromURLParam = (possibleCodeHostType: string): ExternalServiceKind | null => {
    const possibleKind = ExternalServiceKind[possibleCodeHostType.toUpperCase() as ExternalServiceKind]

    return possibleKind ?? null
}

export const isAnyConnectedCodeHosts = (data?: GetCodeHostsResult): boolean => {
    if (!data) {
        return false
    }

    return data.externalServices.nodes.length > 0
}
