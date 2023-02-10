import { mdiBitbucket, mdiGithub, mdiGitlab, mdiAws } from '@mdi/js'

import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'

export const getCodeHostIcon = (codeHostType: ExternalServiceKind): string => {
    switch (codeHostType) {
        case ExternalServiceKind.GITHUB:
            return mdiGithub
        case ExternalServiceKind.GITLAB:
            return mdiGitlab
        case ExternalServiceKind.BITBUCKETCLOUD:
            return mdiBitbucket
        case ExternalServiceKind.AWSCODECOMMIT:
            return mdiAws

        default:
            // TODO: Add support for other code host
            return ''
    }
}

export const getCodeHostName = (codeHostType: ExternalServiceKind): string => {
    switch (codeHostType) {
        case ExternalServiceKind.GITHUB:
            return 'GitHub'
        case ExternalServiceKind.GITLAB:
            return 'GitLab'
        case ExternalServiceKind.BITBUCKETCLOUD:
            return 'BitBucket.org'
        case ExternalServiceKind.AWSCODECOMMIT:
            return 'AWS Code Commit'

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
