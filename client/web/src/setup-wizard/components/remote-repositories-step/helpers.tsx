import type { FC } from 'react'

import { mdiAws, mdiBitbucket, mdiGit, mdiGithub, mdiGitlab, mdiMicrosoftAzure } from '@mdi/js'
import type { MdiReactIconComponentType } from 'mdi-react'

import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'
import { Icon, type IconProps } from '@sourcegraph/wildcard'

import { GerritIcon } from '../../../components/externalServices/GerritIcon'
import type { GetCodeHostsResult } from '../../../graphql-operations'

type CodeHostIconProps = IconProps & {
    codeHostType: ExternalServiceKind | null
}

export const CodeHostIcon: FC<CodeHostIconProps> = props => {
    const { codeHostType, ...attributes } = props
    const codeHostIconPath = getCodeHostIconPath(codeHostType)

    if (codeHostIconPath) {
        return <Icon svgPath={codeHostIconPath} {...attributes} />
    }

    if (codeHostType === ExternalServiceKind.GERRIT) {
        return <GerritIcon size={16} {...(attributes as unknown as MdiReactIconComponentType)} />
    }

    return null
}

export const getCodeHostIconPath = (codeHostType: ExternalServiceKind | null): string | null => {
    switch (codeHostType) {
        case ExternalServiceKind.GITHUB: {
            return mdiGithub
        }
        case ExternalServiceKind.BITBUCKETCLOUD: {
            return mdiBitbucket
        }
        case ExternalServiceKind.BITBUCKETSERVER: {
            return mdiBitbucket
        }
        case ExternalServiceKind.GITLAB: {
            return mdiGitlab
        }
        case ExternalServiceKind.GITOLITE: {
            return mdiGit
        }
        case ExternalServiceKind.AWSCODECOMMIT: {
            return mdiAws
        }
        case ExternalServiceKind.AZUREDEVOPS: {
            return mdiMicrosoftAzure
        }
        default: {
            // TODO: Add support for other code host
            return null
        }
    }
}

export const getCodeHostName = (codeHostType: ExternalServiceKind | null): string => {
    switch (codeHostType) {
        case ExternalServiceKind.GITHUB: {
            return 'GitHub'
        }
        case ExternalServiceKind.GITLAB: {
            return 'GitLab'
        }
        case ExternalServiceKind.BITBUCKETCLOUD: {
            return 'BitBucket.org'
        }
        case ExternalServiceKind.BITBUCKETSERVER: {
            return 'BitBucket Server'
        }
        case ExternalServiceKind.AWSCODECOMMIT: {
            return 'AWS Code Commit'
        }
        case ExternalServiceKind.GITOLITE: {
            return 'Gitolite'
        }
        case ExternalServiceKind.GERRIT: {
            return 'Gerrit'
        }
        case ExternalServiceKind.AZUREDEVOPS: {
            return 'Azure DevOps'
        }
        default: {
            // TODO: Add support for other code host
            return 'Unknown'
        }
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

export const getNextButtonLabel = (data?: GetCodeHostsResult): string => {
    // Fallback for the state when we have no data,
    // While we are loading data the next button is disabled
    if (!data) {
        return 'Loading'
    }

    const otherKindExternalServices = data.externalServices.nodes.filter(
        service => service.kind === ExternalServiceKind.OTHER
    )

    const nonOtherExternalServices = data.externalServices.nodes.filter(
        service => service.kind !== ExternalServiceKind.OTHER
    )

    // This means that we got some "Other" externals services (which are used for
    // local repositories' setup) and have no common external services, meaning that we
    // have no connected code host on remote code host step but got some repositories
    // from local repositories step so next button should have "Skip" label
    if (otherKindExternalServices.length > 0 && nonOtherExternalServices.length === 0) {
        return 'Skip'
    }

    return 'Next'
}

export const getNextButtonLogEvent = (data?: GetCodeHostsResult): string | null => {
    if (!data) {
        return null
    }

    const otherKindExternalServices = data.externalServices.nodes.filter(
        service => service.kind === ExternalServiceKind.OTHER
    )

    const nonOtherExternalServices = data.externalServices.nodes.filter(
        service => service.kind !== ExternalServiceKind.OTHER
    )

    if (otherKindExternalServices.length > 0 && nonOtherExternalServices.length === 0) {
        return 'SetupWizardSkippedAddRemoteCode'
    }

    return null
}
