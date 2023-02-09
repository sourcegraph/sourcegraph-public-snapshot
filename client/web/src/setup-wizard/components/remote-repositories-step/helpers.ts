import { mdiBitbucket, mdiGithub, mdiGitlab } from '@mdi/js';

import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations';

export const getCodeHostIcon = (codeHostType: ExternalServiceKind): string => {
    switch (codeHostType) {
        case ExternalServiceKind.GITHUB:
            return mdiGithub
        case ExternalServiceKind.GITLAB:
            return mdiGitlab
        case ExternalServiceKind.BITBUCKETCLOUD:
            return mdiBitbucket

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

        default:
            // TODO: Add support for other code host
            return 'Unknown'
    }
}
