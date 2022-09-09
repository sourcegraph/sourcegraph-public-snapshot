import React, { useCallback, useMemo } from 'react'

import { upperFirst, toLower } from 'lodash'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import ExportIcon from 'mdi-react/ExportIcon'
import GithubIcon from 'mdi-react/GithubIcon'
import GitlabIcon from 'mdi-react/GitlabIcon'
import { merge, of } from 'rxjs'
import { catchError } from 'rxjs/operators'

import { asError, ErrorLike, isErrorLike } from '@sourcegraph/common'
import { Position, Range } from '@sourcegraph/extension-api-types'
import { SimpleActionItem } from '@sourcegraph/shared/src/actions/SimpleActionItem'
import { PhabricatorIcon } from '@sourcegraph/shared/src/components/icons' // TODO: Switch mdi icon
import { RevisionSpec, FileSpec } from '@sourcegraph/shared/src/util/url'
import { useObservable, Icon, Link, Tooltip, ButtonLinkProps } from '@sourcegraph/wildcard'

import { ExternalLinkFields, RepositoryFields, ExternalServiceKind } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchFileExternalLinks } from '../backend'
import { RepoHeaderActionAnchor, RepoHeaderActionMenuLink } from '../components/RepoHeaderActions'
import { RepoHeaderContext } from '../RepoHeader'

interface Props extends RevisionSpec, Partial<FileSpec> {
    repo?: Pick<RepositoryFields, 'name' | 'defaultBranch' | 'externalURLs'> | null
    filePath?: string
    commitRange?: string
    position?: Position
    range?: Range

    externalLinks?: ExternalLinkFields[]

    fetchFileExternalLinks: typeof fetchFileExternalLinks

    actionType?: 'nav' | 'dropdown'

    source?: 'repoHeader' | 'actionItemsBar'
}

/**
 * A repository header action that goes to the corresponding URL on an external code host.
 */
export const GoToCodeHostAction: React.FunctionComponent<
    React.PropsWithChildren<Props & RepoHeaderContext>
> = props => {
    const { repo, revision, filePath } = props

    /**
     * The external links for the current file/dir, or undefined while loading, null while not
     * needed (because not viewing a file/dir), or an error.
     */
    const fileExternalLinksOrError = useObservable<ExternalLinkFields[] | null | undefined | ErrorLike>(
        useMemo(() => {
            if (!repo || !filePath) {
                return of(null)
            }
            return merge(
                of(undefined),
                fetchFileExternalLinks({ repoName: repo.name, revision, filePath }).pipe(
                    catchError(error => [asError(error)])
                )
            )
        }, [repo, revision, filePath])
    )

    const onClick = useCallback(() => eventLogger.log('GoToCodeHostClicked'), [])

    // If the default branch is undefined, set to HEAD
    const defaultBranch =
        (!isErrorLike(props.repo) && props.repo && props.repo.defaultBranch && props.repo.defaultBranch.displayName) ||
        'HEAD'

    // If neither repo or file can be loaded, return null, which will hide all code host icons
    if (!props.repo || isErrorLike(fileExternalLinksOrError)) {
        return null
    }

    let externalURLs: ExternalLinkFields[]
    if (props.externalLinks && props.externalLinks.length > 0) {
        externalURLs = props.externalLinks
    } else if (
        fileExternalLinksOrError === null ||
        fileExternalLinksOrError === undefined ||
        isErrorLike(fileExternalLinksOrError) ||
        fileExternalLinksOrError.length === 0
    ) {
        // If the external link for the more specific resource within the repository is loading or errored, use the
        // repository external link.
        externalURLs = props.repo.externalURLs
    } else {
        externalURLs = fileExternalLinksOrError
    }
    if (externalURLs.length === 0) {
        return null
    }

    // Only show the first external link for now.
    const externalURL = externalURLs[0]

    const { displayName, icon } = serviceKindDisplayNameAndIcon(externalURL.serviceKind)
    const exportIcon = icon || ExportIcon

    // Extract url to add branch, line numbers or commit range.
    let url = externalURL.url
    if (
        externalURL.serviceKind === ExternalServiceKind.GITHUB ||
        externalURL.serviceKind === ExternalServiceKind.GITLAB
    ) {
        // If in a branch, add branch path to the code host URL.
        if (props.revision && props.revision !== defaultBranch && !fileExternalLinksOrError) {
            url += `/tree/${props.revision}`
        }
        // If showing a comparison, add comparison specifier to the code host URL.
        if (props.commitRange) {
            url += `/compare/${props.commitRange.replace(/^\.{3}/, 'HEAD...').replace(/\.{3}$/, '...HEAD')}`
        }
        // Add range or position path to the code host URL.
        if (props.range) {
            const rangeEndPrefix = externalURL.serviceKind === ExternalServiceKind.GITLAB ? '' : 'L'
            url += `#L${props.range.start.line}-${rangeEndPrefix}${props.range.end.line}`
        } else if (props.position) {
            url += `#L${props.position.line}`
        }
    }

    const TARGET_ID = 'go-to-code-host'

    const descriptiveText = `View on ${displayName}`

    const commonProps: Partial<ButtonLinkProps> = {
        to: url,
        target: '_blank',
        rel: 'noopener noreferrer',
        id: TARGET_ID,
        onClick,
        onAuxClick: onClick,
        className: 'test-go-to-code-host',
        'aria-label': descriptiveText,
    }

    if (props.source === 'actionItemsBar') {
        return (
            <SimpleActionItem tooltip={descriptiveText} {...commonProps}>
                <Icon as={exportIcon} aria-hidden={true} />
            </SimpleActionItem>
        )
    }

    // Don't show browser extension popover on small screens
    if (props.actionType === 'dropdown') {
        return (
            <RepoHeaderActionMenuLink
                as={Link}
                className="test-go-to-code-host"
                // empty href is OK because we always set tabindex=0
                to={url}
                target="_blank"
                file={true}
                rel="noopener noreferrer"
                id={TARGET_ID}
                onClick={onClick}
                onAuxClick={onClick}
            >
                <Icon as={exportIcon} aria-hidden={true} />
                <span>{descriptiveText}</span>
            </RepoHeaderActionMenuLink>
        )
    }

    return (
        <Tooltip content={descriptiveText}>
            <RepoHeaderActionAnchor {...commonProps}>
                <Icon as={exportIcon} aria-hidden={true} />
            </RepoHeaderActionAnchor>
        </Tooltip>
    )
}

export function serviceKindDisplayNameAndIcon(
    serviceKind: ExternalServiceKind | null
): { displayName: string; icon?: React.ComponentType<{ className?: string }> } {
    if (!serviceKind) {
        return { displayName: 'code host' }
    }

    switch (serviceKind) {
        case ExternalServiceKind.GITHUB:
            return { displayName: 'GitHub', icon: GithubIcon }
        case ExternalServiceKind.GITLAB:
            return { displayName: 'GitLab', icon: GitlabIcon }
        case ExternalServiceKind.BITBUCKETSERVER:
            return { displayName: 'Bitbucket Server', icon: BitbucketIcon }
        case ExternalServiceKind.BITBUCKETCLOUD:
            return { displayName: 'Bitbucket Cloud', icon: BitbucketIcon }
        case ExternalServiceKind.PHABRICATOR:
            return { displayName: 'Phabricator', icon: PhabricatorIcon }
        case ExternalServiceKind.AWSCODECOMMIT:
            return { displayName: 'AWS CodeCommit' }
        default:
            return { displayName: upperFirst(toLower(serviceKind)) }
    }
}
