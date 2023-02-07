import React, { useCallback, useEffect, useMemo, useState } from 'react'

import { mdiChevronDoubleDown, mdiChevronDoubleUp, mdiMenuDown, mdiMenuUp, mdiPuzzleOutline } from '@mdi/js'
import VisuallyHidden from '@reach/visually-hidden'
import classNames from 'classnames'
import { head, last } from 'lodash'
import { useLocation } from 'react-router-dom-v5-compat'
import { BehaviorSubject, from, of } from 'rxjs'
import { distinctUntilChanged, map } from 'rxjs/operators'
import { focusable, FocusableElement } from 'tabbable'
import { Key } from 'ts-key-enum'

import { ContributableMenu } from '@sourcegraph/client-api'
import { LocalStorageSubject } from '@sourcegraph/common'
import { ActionItem } from '@sourcegraph/shared/src/actions/ActionItem'
import { ActionsContainer } from '@sourcegraph/shared/src/actions/ActionsContainer'
import { haveInitialExtensionsLoaded } from '@sourcegraph/shared/src/api/features'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { isSettingsValid } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, ButtonLink, Icon, LoadingSpinner, Tooltip, useObservable } from '@sourcegraph/wildcard'

import { ErrorBoundary } from '../../components/ErrorBoundary'
import { useCarousel } from '../../components/useCarousel'
import { RepositoryFields } from '../../graphql-operations'
import { OpenInEditorActionItem } from '../../open-in-editor/OpenInEditorActionItem'
import { GoToCodeHostAction } from '../../repo/actions/GoToCodeHostAction'
import { ToggleBlameAction } from '../../repo/actions/ToggleBlameAction'
import { fetchFileExternalLinks } from '../../repo/backend'
import { parseBrowserRepoURL } from '../../util/url'

import styles from './ActionItemsBar.module.scss'

const scrollButtonClassName = styles.scroll

function getIconClassName(index: number): string | undefined {
    return (styles as Record<string, string>)[`icon${index % 5}`]
}

function arrowable(element: HTMLElement): FocusableElement[] {
    return focusable(element).filter(
        elm => !elm.classList.contains('disabled') && !elm.classList.contains(scrollButtonClassName)
    )
}

export function useWebActionItems(): Pick<ActionItemsBarProps, 'useActionItemsBar'> &
    Pick<ActionItemsToggleProps, 'useActionItemsToggle'> {
    const toggles = useMemo(() => new LocalStorageSubject('action-items-bar-expanded', true), [])

    const [toggleReference, setToggleReference] = useState<HTMLElement | null>(null)
    const nextToggleReference = useCallback((toggle: HTMLElement) => {
        setToggleReference(toggle)
    }, [])

    const [barReference, setBarReference] = useState<HTMLElement | null>(null)
    const nextBarReference = useCallback((bar: HTMLElement) => {
        setBarReference(bar)
    }, [])

    // Set up keyboard navigation for distant toggle and bar. Remove previous event
    // listeners whenever references change.
    useEffect(() => {
        function onKeyDownToggle(event: KeyboardEvent): void {
            if (event.key === Key.ArrowDown && barReference) {
                const firstBarArrowable = head(arrowable(barReference))
                if (firstBarArrowable) {
                    firstBarArrowable.focus()
                    event.preventDefault()
                }
            }

            if (event.key === Key.ArrowUp && barReference) {
                const lastBarArrowable = last(arrowable(barReference))
                if (lastBarArrowable) {
                    lastBarArrowable.focus()
                    event.preventDefault()
                }
            }
        }

        function onKeyDownBar(event: KeyboardEvent): void {
            if (event.target instanceof HTMLElement && toggleReference && barReference) {
                const arrowableChildren = arrowable(barReference)
                const indexOfTarget = arrowableChildren.indexOf(event.target)

                if (event.key === Key.ArrowDown) {
                    // If this is the last arrowable element, go back to the toggle
                    if (indexOfTarget === arrowableChildren.length - 1) {
                        toggleReference.focus()
                        event.preventDefault()
                        return
                    }

                    const itemToFocus = arrowableChildren[indexOfTarget + 1]
                    if (itemToFocus instanceof HTMLElement) {
                        itemToFocus.focus()
                        event.preventDefault()
                        return
                    }
                }

                if (event.key === Key.ArrowUp) {
                    // If this is the first arrowable element, go back to the toggle
                    if (indexOfTarget === 0) {
                        toggleReference.focus()
                        event.preventDefault()
                        return
                    }

                    const itemToFocus = arrowableChildren[indexOfTarget - 1]
                    if (itemToFocus instanceof HTMLElement) {
                        itemToFocus.focus()
                        event.preventDefault()
                        return
                    }
                }
            }
        }

        toggleReference?.addEventListener('keydown', onKeyDownToggle)
        barReference?.addEventListener('keydown', onKeyDownBar)

        return () => {
            toggleReference?.removeEventListener('keydown', onKeyDownToggle)
            toggleReference?.removeEventListener('keydown', onKeyDownBar)
        }
    }, [toggleReference, barReference])

    const barsReferenceCounts = useMemo(() => new BehaviorSubject(0), [])

    const useActionItemsBar = useCallback(() => {
        // `useActionItemsBar` will be used as a hook
        // eslint-disable-next-line react-hooks/rules-of-hooks
        const isOpen = useObservable(toggles)

        // Let the toggle know it's on the page
        // eslint-disable-next-line react-hooks/rules-of-hooks
        useEffect(() => {
            // Use reference counter so that effect order doesn't matter
            barsReferenceCounts.next(barsReferenceCounts.value + 1)

            return () => barsReferenceCounts.next(barsReferenceCounts.value - 1)
        }, [])

        return { isOpen, barReference: nextBarReference }
    }, [toggles, nextBarReference, barsReferenceCounts])

    const useActionItemsToggle = useCallback(() => {
        // `useActionItemsToggle` will be used as a hook
        // eslint-disable-next-line react-hooks/rules-of-hooks
        const isOpen = useObservable(toggles)

        // eslint-disable-next-line react-hooks/rules-of-hooks
        const toggle = useCallback(() => toggles.next(!isOpen), [isOpen])

        // Only show the action items toggle when the <ActionItemsBar> component is on the page
        // eslint-disable-next-line react-hooks/rules-of-hooks
        const barInPage = !!useObservable(
            // eslint-disable-next-line react-hooks/rules-of-hooks
            useMemo(
                () =>
                    barsReferenceCounts.pipe(
                        map(count => count > 0),
                        distinctUntilChanged()
                    ),
                []
            )
        )

        return { isOpen, toggle, barInPage, toggleReference: nextToggleReference }
    }, [toggles, nextToggleReference, barsReferenceCounts])

    return {
        useActionItemsBar,
        useActionItemsToggle,
    }
}

export interface ActionItemsBarProps extends ExtensionsControllerProps, TelemetryProps, PlatformContextProps {
    repo?: RepositoryFields
    useActionItemsBar: () => { isOpen: boolean | undefined; barReference: React.RefCallback<HTMLElement> }
    source?: 'compare' | 'commit' | 'blob'
}

const actionItemClassName = classNames(
    'd-flex justify-content-center align-items-center text-decoration-none',
    styles.action
)

/**
 * Renders extensions (both migrated to the core workflow and legacy) actions items in the sidebar.
 */
export const ActionItemsBar = React.memo<ActionItemsBarProps>(function ActionItemsBar(props) {
    const { extensionsController, source } = props

    const location = useLocation()
    const { isOpen, barReference } = props.useActionItemsBar()
    const { repoName, rawRevision, filePath, commitRange, position, range } = parseBrowserRepoURL(
        location.pathname + location.search + location.hash
    )

    const { carouselReference, canScrollNegative, canScrollPositive, onNegativeClicked, onPositiveClicked } =
        useCarousel({ direction: 'topToBottom' })

    const haveExtensionsLoaded = useObservable(
        useMemo(
            () =>
                extensionsController !== null ? haveInitialExtensionsLoaded(extensionsController.extHostAPI) : of(true),
            [extensionsController]
        )
    )

    const settingsOrError = useObservable(
        useMemo(() => from(props.platformContext.settings), [props.platformContext.settings])
    )
    const settings =
        settingsOrError !== undefined && isSettingsValid(settingsOrError) ? settingsOrError.final : undefined
    const perforceCodeHostUrlToSwarmUrlMap =
        (settings?.['perforce.codeHostToSwarmMap'] as { [codeHost: string]: string } | undefined) || {}

    if (!isOpen) {
        return <div className={classNames(styles.bar, styles.barCollapsed)} />
    }

    return (
        <div className={classNames('p-0 mr-2 position-relative d-flex flex-column', styles.bar)} ref={barReference}>
            {/* To be clear to users that this isn't an error reported by extensions about e.g. the code they're viewing. */}
            <ErrorBoundary location={location} render={error => <span>Component error: {error.message}</span>}>
                <ActionItemsDivider />
                {canScrollNegative && (
                    <Button
                        className={classNames('p-0 border-0', styles.scroll, styles.listItem)}
                        onClick={onNegativeClicked}
                        tabIndex={-1}
                        variant="link"
                        aria-label="Scroll up"
                    >
                        <Icon aria-hidden={true} svgPath={mdiMenuUp} />
                    </Button>
                )}

                {source !== 'compare' && source !== 'commit' && (
                    <GoToCodeHostAction
                        repo={props.repo}
                        repoName={repoName}
                        // We need a revision to generate code host URLs, if revision isn't available, we use the default branch or HEAD.
                        revision={rawRevision || props.repo?.defaultBranch?.displayName || 'HEAD'}
                        filePath={filePath}
                        commitRange={commitRange}
                        range={range}
                        position={position}
                        perforceCodeHostUrlToSwarmUrlMap={perforceCodeHostUrlToSwarmUrlMap}
                        fetchFileExternalLinks={fetchFileExternalLinks}
                        actionType="nav"
                        source="actionItemsBar"
                    />
                )}

                {source === 'blob' && (
                    <>
                        <ToggleBlameAction actionType="nav" source="actionItemsBar" />
                        {window.context.isAuthenticatedUser && (
                            <OpenInEditorActionItem
                                platformContext={props.platformContext}
                                externalServiceType={props.repo?.externalRepository?.serviceType}
                                actionType="nav"
                                source="actionItemsBar"
                            />
                        )}
                    </>
                )}

                {extensionsController !== null ? (
                    <ActionsContainer
                        menu={ContributableMenu.EditorTitle}
                        returnInactiveMenuItems={true}
                        extensionsController={extensionsController}
                        empty={null}
                        location={location}
                        platformContext={props.platformContext}
                        telemetryService={props.telemetryService}
                    >
                        {items => (
                            <ul className={classNames('list-unstyled m-0', styles.list)} ref={carouselReference}>
                                {items.map((item, index) => {
                                    const hasIconURL = !!item.action.actionItem?.iconURL
                                    const className = classNames(
                                        actionItemClassName,
                                        !hasIconURL &&
                                            classNames(styles.actionNoIcon, getIconClassName(index), 'text-sm')
                                    )
                                    const inactiveClassName = classNames(
                                        styles.actionInactive,
                                        !hasIconURL && styles.actionNoIconInactive
                                    )
                                    const listItemClassName = classNames(
                                        styles.listItem,
                                        index !== items.length - 1 && 'mb-1'
                                    )

                                    const dataContent = !hasIconURL ? item.action.category?.slice(0, 1) : undefined

                                    return (
                                        <li key={item.action.id} className={listItemClassName}>
                                            <ActionItem
                                                {...props}
                                                {...item}
                                                location={location}
                                                extensionsController={extensionsController}
                                                className={className}
                                                dataContent={dataContent}
                                                variant="actionItem"
                                                iconClassName={styles.icon}
                                                pressedClassName={styles.actionPressed}
                                                inactiveClassName={inactiveClassName}
                                                hideLabel={true}
                                                tabIndex={-1}
                                                hideExternalLinkIcon={true}
                                                disabledDuringExecution={true}
                                            />
                                        </li>
                                    )
                                })}
                            </ul>
                        )}
                    </ActionsContainer>
                ) : null}
                {canScrollPositive && (
                    <Button
                        className={classNames('p-0 border-0', styles.scroll, styles.listItem)}
                        onClick={onPositiveClicked}
                        tabIndex={-1}
                        variant="link"
                        aria-label="Scroll down"
                    >
                        <Icon aria-hidden={true} svgPath={mdiMenuDown} />
                    </Button>
                )}
                {haveExtensionsLoaded && <ActionItemsDivider />}
            </ErrorBoundary>
        </div>
    )
})

export interface ActionItemsToggleProps extends ExtensionsControllerProps<'extHostAPI'> {
    useActionItemsToggle: () => {
        isOpen: boolean | undefined
        toggle: () => void
        toggleReference: React.RefCallback<HTMLElement>
        barInPage: boolean
    }
    className?: string
}

export const ActionItemsToggle: React.FunctionComponent<React.PropsWithChildren<ActionItemsToggleProps>> = ({
    useActionItemsToggle,
    extensionsController,
    className,
}) => {
    const panelName = extensionsController !== null && window.context.enableLegacyExtensions ? 'extensions' : 'actions'

    const { isOpen, toggle, toggleReference, barInPage } = useActionItemsToggle()

    const haveExtensionsLoaded = useObservable(
        useMemo(
            () =>
                extensionsController !== null ? haveInitialExtensionsLoaded(extensionsController.extHostAPI) : of(true),
            [extensionsController]
        )
    )

    return barInPage ? (
        <>
            <li className={styles.dividerVertical} />
            <li className={classNames('nav-item mr-2', className)}>
                <div className={classNames(styles.toggleContainer, isOpen && styles.toggleContainerOpen)}>
                    <Tooltip content={`${isOpen ? 'Close' : 'Open'} ${panelName} panel`}>
                        <ButtonLink
                            aria-label={
                                isOpen
                                    ? `Close ${panelName} panel. Press the down arrow key to enter the ${panelName} panel.`
                                    : `Open ${panelName} panel`
                            }
                            className={classNames(actionItemClassName, styles.auxIcon, styles.actionToggle)}
                            onSelect={toggle}
                            ref={toggleReference}
                        >
                            {!haveExtensionsLoaded ? (
                                <LoadingSpinner />
                            ) : isOpen ? (
                                <Icon aria-hidden={true} svgPath={mdiChevronDoubleUp} />
                            ) : (
                                <Icon
                                    aria-hidden={true}
                                    svgPath={
                                        window.context.enableLegacyExtensions ? mdiPuzzleOutline : mdiChevronDoubleDown
                                    }
                                />
                            )}
                            {haveExtensionsLoaded && <VisuallyHidden>Down arrow to enter</VisuallyHidden>}
                        </ButtonLink>
                    </Tooltip>
                </div>
            </li>
        </>
    ) : null
}

const ActionItemsDivider: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => <div className={classNames('position-relative rounded-sm d-flex', styles.dividerHorizontal, className)} />
