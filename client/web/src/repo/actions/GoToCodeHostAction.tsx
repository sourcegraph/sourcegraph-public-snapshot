import { Position, Range } from '@sourcegraph/extension-api-types'
import { upperFirst } from 'lodash'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import ExportIcon from 'mdi-react/ExportIcon'
import GithubIcon from 'mdi-react/GithubIcon'
import PlusThickIcon from 'mdi-react/PlusThickIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { merge, Observable, of } from 'rxjs'
import { catchError, distinctUntilChanged, startWith, switchMap } from 'rxjs/operators'
import { PhabricatorIcon } from '../../../../shared/src/components/icons' // TODO: Switch mdi icon
import { ButtonLink } from '../../../../shared/src/components/LinkOrButton'
import * as GQL from '../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { fetchFileExternalLinks } from '../backend'
import { RevisionSpec, FileSpec } from '../../../../shared/src/util/url'
import { ExternalLinkFields } from '../../graphql-operations'
import { useEventObservable, useObservable } from '../../../../shared/src/util/useObservable'
import GitlabIcon from 'mdi-react/GitlabIcon'
import { SourcegraphIcon } from '../../auth/icons'
import { eventLogger } from '../../tracking/eventLogger'
import { PopoverContainer } from '../../components/PopoverContainer'

interface GoToCodeHostPopoverProps {
    canShowPopover: boolean
    onPopoverDismissed: () => void
}

interface Props extends RevisionSpec, Partial<FileSpec>, GoToCodeHostPopoverProps {
    repo?: Pick<GQL.IRepository, 'name' | 'defaultBranch' | 'externalURLs'> | null
    filePath?: string
    commitRange?: string
    position?: Position
    range?: Range

    externalLinks?: ExternalLinkFields[]

    browserExtensionInstalled: Observable<boolean | { platform: unknown }>
    fetchFileExternalLinks: typeof fetchFileExternalLinks
}

const HAS_DISMISSED_POPUP_KEY = 'has-dismissed-browser-ext-popup'

/**
 * A repository header action that goes to the corresponding URL on an external code host.
 */
export const GoToCodeHostAction: React.FunctionComponent<Props> = props => {
    const [modalOpen, setModalOpen] = useState(false)
    /**
     * TODO: only set popover open when "canShowPopover" is true AND "hasDismissedPopover" is false
     */

    const { onPopoverDismissed } = props

    const isExtensionInstalled = useObservable(props.browserExtensionInstalled)

    const [hasDissmissedPopup, setHasDismissedPopup] = useState(false)

    const hijackLink = !isExtensionInstalled && !hasDissmissedPopup

    useEffect(() => {
        setHasDismissedPopup(localStorage.getItem(HAS_DISMISSED_POPUP_KEY) === 'true')
    }, [])

    /**
     * The external links for the current file/dir, or undefined while loading, null while not
     * needed (because not viewing a file/dir), or an error.
     */
    const [nextComponentUpdate, fileExternalLinksOrError] = useEventObservable<
        Props,
        ExternalLinkFields[] | null | undefined | ErrorLike
    >(
        useCallback(
            componentUpdates =>
                componentUpdates.pipe(
                    startWith(props),
                    distinctUntilChanged(
                        (a, b) => a.repo === b.repo && a.revision === b.revision && a.filePath === b.filePath
                    ),
                    switchMap(({ repo, revision, filePath }) => {
                        if (!repo || !filePath) {
                            return of(null)
                        }
                        return merge(
                            of(undefined),
                            fetchFileExternalLinks({ repoName: repo.name, revision, filePath }).pipe(
                                catchError(error => [asError(error)])
                            )
                        )
                    })
                ),
            // Pass latest props in `useEffect`, don't want to create new subscriptions on each render
            // eslint-disable-next-line react-hooks/exhaustive-deps
            []
        )
    )

    useEffect(() => {
        nextComponentUpdate(props)
    }, [props, nextComponentUpdate])

    /** This is a hard rejection. Never ask the user again. */
    const onRejection = useCallback(() => {
        localStorage.setItem(HAS_DISMISSED_POPUP_KEY, 'true')
        setHasDismissedPopup(true)
        onPopoverDismissed()

        eventLogger.log('BrowserExtensionPopupRejected')
    }, [onPopoverDismissed])

    /** This is a soft rejection. Called when user clicks 'Remind me later', ESC, or outside of the modal body */
    const onClose = useCallback(() => {
        onPopoverDismissed()
        setModalOpen(false)
        eventLogger.log('BrowserExtensionPopupClosed')
    }, [onPopoverDismissed])

    /** The user is likely to install the browser extension at this point, so don't show it again. */
    const onClickInstall = useCallback(() => {
        localStorage.setItem(HAS_DISMISSED_POPUP_KEY, 'true')
        setHasDismissedPopup(true)
        onPopoverDismissed()

        eventLogger.log('BrowserExtensionPopupClickedInstall')
    }, [onPopoverDismissed])

    const onSelect = useCallback(() => {
        if (modalOpen) {
            setModalOpen(false)
            return
        }

        if (hijackLink) {
            setModalOpen(true)
        }
    }, [hijackLink, modalOpen])

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

    const { displayName, icon } = serviceTypeDisplayNameAndIcon(externalURL.serviceType)
    const Icon = icon || ExportIcon

    // Extract url to add branch, line numbers or commit range.
    let url = externalURL.url
    if (externalURL.serviceType === 'github' || externalURL.serviceType === 'gitlab') {
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
            const rangeEndPrefix = externalURL.serviceType === 'gitlab' ? '' : 'L'
            url += `#L${props.range.start.line}-${rangeEndPrefix}${props.range.end.line}`
        } else if (props.position) {
            url += `#L${props.position.line}`
        }
    }

    const TARGET_ID = 'go-to-code-host'

    return (
        <>
            <ButtonLink
                className="nav-link test-go-to-code-host"
                // empty href is OK because we always set tabindex=0
                to={hijackLink ? '' : url}
                target="_self"
                data-tooltip={`View on ${displayName}`}
                onSelect={onSelect}
                id={TARGET_ID}
            >
                <Icon className="icon-inline" />
            </ButtonLink>

            {modalOpen && (
                <CodeHostExtensionPopover
                    url={url}
                    serviceType={externalURL.serviceType}
                    onClose={onClose}
                    onRejection={onRejection}
                    onClickInstall={onClickInstall}
                    targetID={TARGET_ID}
                />
            )}
        </>
    )
}

interface CodeHostExtensionPopoverProps {
    url: string
    serviceType: string | null
    onClose: () => void
    onRejection: () => void
    onClickInstall: () => void
    targetID: string
}

export const CodeHostExtensionPopover: React.FunctionComponent<CodeHostExtensionPopoverProps> = ({
    url,
    serviceType,
    onClose,
    onRejection,
    onClickInstall,
    targetID,
}) => {
    const { displayName, icon } = serviceTypeDisplayNameAndIcon(serviceType)
    const Icon = icon || ExportIcon

    return (
        <PopoverContainer
            onClose={onClose}
            targetID={targetID}
            popperOptions={useMemo(
                () => ({
                    placement: 'bottom-start' as const,
                    modifiers: [
                        {
                            name: 'offset',
                            options: {
                                offset: [64, 4],
                            },
                        },
                    ],
                }),
                []
            )}
        >
            {modalBodyReference => (
                <div
                    ref={modalBodyReference as React.MutableRefObject<HTMLDivElement>}
                    className="extension-permission-modal p-4 web-content text-wrap border shadow"
                >
                    <h3 className="mb-0">Take Sourcegraph's code intelligence to {displayName}!</h3>
                    <p className="py-3">
                        Install Sourcegraph browser extension to get code intelligence while browsing files and reading
                        PRs on {displayName}.
                    </p>

                    <div className="mx-auto code-host-action__graphic-container d-flex justify-content-between align-items-center">
                        <SourcegraphIcon size={48} />
                        <PlusThickIcon size={20} className="code-host-action__plus-icon" />
                        <Icon size={56} />
                    </div>

                    <div className="d-flex justify-content-end">
                        <ButtonLink className="btn btn-outline-secondary mr-2" onSelect={onRejection} to={url}>
                            No, thanks
                        </ButtonLink>

                        <ButtonLink className="btn btn-outline-secondary mr-2" onSelect={onClose} to={url}>
                            Remind me later
                        </ButtonLink>

                        <ButtonLink
                            className="btn btn-primary mr-2"
                            onSelect={onClickInstall}
                            to="/help/integration/browser_extension"
                        >
                            Install browser extension
                        </ButtonLink>
                    </div>
                </div>
            )}
        </PopoverContainer>
    )
}

function serviceTypeDisplayNameAndIcon(
    serviceType: string | null
): { displayName: string; icon?: React.ComponentType<{ className?: string }> } {
    switch (serviceType) {
        case 'github':
            return { displayName: 'GitHub', icon: GithubIcon }
        case 'gitlab':
            return { displayName: 'GitLab', icon: GitlabIcon }
        case 'bitbucketServer':
            // TODO: Why is bitbucketServer (correctly) camelCase but
            // awscodecommit is (correctly) lowercase? Why is serviceType
            // not type-checked for validity?
            return { displayName: 'Bitbucket Server', icon: BitbucketIcon }
        case 'phabricator':
            return { displayName: 'Phabricator', icon: PhabricatorIcon }
        case 'awscodecommit':
            return { displayName: 'AWS CodeCommit' }
    }
    return { displayName: serviceType ? upperFirst(serviceType) : 'code host' }
}
