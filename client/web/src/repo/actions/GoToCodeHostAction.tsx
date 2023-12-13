import React, { useCallback, useMemo } from 'react'

import { toLower, upperFirst } from 'lodash'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import ExportIcon from 'mdi-react/ExportIcon'
import GithubIcon from 'mdi-react/GithubIcon'
import GitlabIcon from 'mdi-react/GitlabIcon'
import { merge, of } from 'rxjs'
import { catchError } from 'rxjs/operators'

import { asError, type ErrorLike, isErrorLike, logger } from '@sourcegraph/common'
import type { Position, Range } from '@sourcegraph/extension-api-types'
import { SimpleActionItem } from '@sourcegraph/shared/src/actions/SimpleActionItem'
// TODO: Switch mdi icon
import { HelixSwarmIcon, PhabricatorIcon } from '@sourcegraph/shared/src/components/icons'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { FileSpec, RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { type ButtonLinkProps, Icon, Link, Tooltip, useObservable } from '@sourcegraph/wildcard'

import { type ExternalLinkFields, ExternalServiceKind, type RepositoryFields } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchCommitMessage, fetchFileExternalLinks } from '../backend'
import { RepoHeaderActionAnchor, RepoHeaderActionMenuLink } from '../components/RepoHeaderActions'
import type { RepoHeaderContext } from '../RepoHeader'

interface Props extends RevisionSpec, Partial<FileSpec>, TelemetryV2Props {
    repo?: Pick<RepositoryFields, 'name' | 'defaultBranch' | 'externalURLs' | 'externalRepository'> | null
    filePath?: string
    commitRange?: string
    range?: Range
    position?: Position

    externalLinks?: ExternalLinkFields[]

    perforceCodeHostUrlToSwarmUrlMap: { [key: string]: string }

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
    const { repo, revision, filePath, telemetryRecorder } = props

    const serviceType = props.repo?.externalRepository?.serviceType

    // The external links for the current file/dir, or undefined while loading,
    // null while not needed (because not viewing a file/dir), or an error.
    const fileExternalLinksOrError = useObservable<ExternalLinkFields[] | null | undefined | ErrorLike>(
        useMemo(() => {
            if (!repo || !filePath || serviceType === 'perforce') {
                return of(null)
            }
            return merge(
                of(undefined),
                fetchFileExternalLinks({ repoName: repo.name, revision, filePath }).pipe(
                    catchError(error => [asError(error)])
                )
            )
        }, [repo, filePath, serviceType, revision])
    )

    // The commit message, or null if not needed, undefined while loading, also null on error.
    const perforceCommitMessage = useObservable<string | null | undefined>(
        useMemo(() => {
            if (
                serviceType !== 'perforce' ||
                !Object.keys(props.perforceCodeHostUrlToSwarmUrlMap).length ||
                !props.repo
            ) {
                return of(null)
            }
            return merge(
                of(undefined),
                fetchCommitMessage({ repoName: props.repoName, revision: props.revision }).pipe(
                    catchError(error => {
                        logger.error('Getting commit message failed', error)
                        return [null]
                    })
                )
            )
        }, [serviceType, props.perforceCodeHostUrlToSwarmUrlMap, props.repo, props.repoName, props.revision])
    )

    const onClick = useCallback(() => {
        telemetryRecorder.recordEvent('goToCodeHost', 'clicked')
        eventLogger.log('GoToCodeHostClicked')
    }, [telemetryRecorder])

    // If the default branch is undefined, set to HEAD
    const defaultBranch =
        (!isErrorLike(props.repo) && props.repo && props.repo.defaultBranch && props.repo.defaultBranch.displayName) ||
        'HEAD'

    // If there's no repo or no file / commit message, return null to hide all code host icons
    if (!props.repo || (isErrorLike(fileExternalLinksOrError) && !perforceCommitMessage)) {
        return null
    }

    const [serviceKind, url] =
        serviceType === 'perforce'
            ? getPerforceServiceKindAndSwarmUrl(
                  props.perforceCodeHostUrlToSwarmUrlMap,
                  props.repo.externalRepository.serviceID,
                  props.repoName,
                  revision,
                  perforceCommitMessage,
                  filePath
              )
            : getServiceKindAndGitUrl(
                  props.externalLinks,
                  props.repo.externalURLs,
                  fileExternalLinksOrError,
                  revision,
                  defaultBranch,
                  props.commitRange,
                  props.range,
                  props.position
              )
    if (!serviceKind || !url) {
        return null
    }

    const { displayName, icon } = serviceKindDisplayNameAndIcon(serviceKind)
    const exportIcon = icon || ExportIcon

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

function getServiceKindAndGitUrl(
    externalLinks: ExternalLinkFields[] | undefined,
    repoExternalURLs: RepositoryFields['externalURLs'] | undefined,
    fileExternalLinksOrError: ExternalLinkFields[] | undefined | null | ErrorLike,
    revision: string,
    defaultBranch: string,
    commitRange: string | undefined,
    range: Range | undefined,
    position: Position | undefined
): [ExternalServiceKind | null, string | null] {
    const externalURLs = getGitExternalURLs(externalLinks, repoExternalURLs, fileExternalLinksOrError)

    if (!externalURLs || externalURLs.length === 0) {
        return [null, null]
    }

    // Only show the first external link for now.
    const serviceKind = externalURLs[0].serviceKind

    let gitUrl = externalURLs[0].url
    if (serviceKind === ExternalServiceKind.GITHUB || serviceKind === ExternalServiceKind.GITLAB) {
        // If in a branch, add branch path to the code host URL.
        if (revision && revision !== defaultBranch && !fileExternalLinksOrError) {
            gitUrl += `/tree/${revision}`
        }
        // If showing a comparison, add comparison specifier to the code host URL.
        if (commitRange) {
            gitUrl += `/compare/${commitRange.replace(/^\.{3}/, 'HEAD...').replace(/\.{3}$/, '...HEAD')}`
        }
        // Add range or position path to the code host URL.
        if (range) {
            const rangeEndPrefix = serviceKind === ExternalServiceKind.GITLAB ? '' : 'L'
            gitUrl += `#L${range.start.line}-${rangeEndPrefix}${range.end.line}`
        } else if (position) {
            gitUrl += `#L${position.line}`
        }
    }
    return [serviceKind, gitUrl]
}

function getGitExternalURLs(
    externalLinks: ExternalLinkFields[] | undefined,
    repoExternalURLs: RepositoryFields['externalURLs'] | undefined,
    fileExternalLinksOrError: ExternalLinkFields[] | undefined | null | ErrorLike
): ExternalLinkFields[] | undefined {
    if (externalLinks && externalLinks.length > 0) {
        return externalLinks
    }
    if (
        fileExternalLinksOrError === null ||
        fileExternalLinksOrError === undefined ||
        isErrorLike(fileExternalLinksOrError) ||
        fileExternalLinksOrError.length === 0
    ) {
        // If the external link for the more specific resource within the repository is loading or errored, use the
        // repository external link.
        return repoExternalURLs
    }

    return fileExternalLinksOrError
}

/**
 * @param perforceCodeHostUrlToSwarmUrlMap Keys should have no prefix and should not end with a slash. Like "perforce.company.com:1666"
 * Values should look like "https://swarm.company.com/", with a slash at the end.
 * @param serviceID is the Perforce hostname, like "perforce.company.com:1666"
 * @param repoName is like "some-repo-name", probably always the depot name without the slashes - TODO: Use this once we figure out the URL format
 * @param revision is the branch name, like "main" - TODO: Use this once we figure out the URL format
 * @param commitMessage should be like "Test\n[git-p4: depot-paths = \"//some-depot-path/\": change = 91512]". Only the end is used.
 * @param filePath is like "test/1.js" - TODO: Use this once we figure out the URL format
 */
function getPerforceServiceKindAndSwarmUrl(
    perforceCodeHostUrlToSwarmUrlMap: { [key: string]: string },
    serviceID: string,
    repoName: string,
    revision: string,
    commitMessage: string | undefined | null,
    filePath: string | undefined
): [ExternalServiceKind | null, string | null] {
    if (!commitMessage) {
        return [null, null]
    }
    if (!Object.keys(perforceCodeHostUrlToSwarmUrlMap).includes(serviceID)) {
        return [null, null]
    }
    const changelistNumber = getPerforceChangelistNumberFromCommitMessage(commitMessage)
    if (!changelistNumber) {
        return [null, null]
    }
    return [ExternalServiceKind.PERFORCE, perforceCodeHostUrlToSwarmUrlMap[serviceID] + changelistNumber]
}

function getPerforceChangelistNumberFromCommitMessage(commitMessage: string): string | null {
    const changeIndex = commitMessage.lastIndexOf('change = ')
    if (changeIndex === -1) {
        return null
    }
    return commitMessage.slice(changeIndex + 9, -1)
}

export function serviceKindDisplayNameAndIcon(serviceKind: ExternalServiceKind | null): {
    displayName: string
    icon?: React.ComponentType<{ className?: string }>
} {
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
        case ExternalServiceKind.PERFORCE:
            return { displayName: 'Swarm', icon: HelixSwarmIcon }
        case ExternalServiceKind.PHABRICATOR:
            return { displayName: 'Phabricator', icon: PhabricatorIcon }
        case ExternalServiceKind.AWSCODECOMMIT:
            return { displayName: 'AWS CodeCommit' }
        default:
            return { displayName: upperFirst(toLower(serviceKind)) }
    }
}
