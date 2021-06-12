import classNames from 'classnames'
import * as H from 'history'
import { head, last } from 'lodash'
import ChevronDoubleUpIcon from 'mdi-react/ChevronDoubleUpIcon'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import PlusIcon from 'mdi-react/PlusIcon'
import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { BehaviorSubject } from 'rxjs'
import { distinctUntilChanged, map } from 'rxjs/operators'
import { focusable, FocusableElement } from 'tabbable'
import { Key } from 'ts-key-enum'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ActionItem } from '@sourcegraph/shared/src/actions/ActionItem'
import { ActionsContainer } from '@sourcegraph/shared/src/actions/ActionsContainer'
import { haveInitialExtensionsLoaded } from '@sourcegraph/shared/src/api/features'
import { ContributableMenu } from '@sourcegraph/shared/src/api/protocol'
import { ButtonLink } from '@sourcegraph/shared/src/components/LinkOrButton'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LocalStorageSubject } from '@sourcegraph/shared/src/util/LocalStorageSubject'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'

import { ErrorBoundary } from '../../components/ErrorBoundary'
import { useCarousel } from '../../components/useCarousel'

const scrollButtonClassName = 'action-items__scroll'

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

export interface ActionItemsBarProps extends ExtensionsControllerProps, PlatformContextProps, TelemetryProps {
    useActionItemsBar: () => { isOpen: boolean | undefined; barReference: React.RefCallback<HTMLElement> }
    location: H.Location
}

export interface ActionItemsToggleProps extends ExtensionsControllerProps<'extHostAPI'> {
    useActionItemsToggle: () => {
        isOpen: boolean | undefined
        toggle: () => void
        toggleReference: React.RefCallback<HTMLElement>
        barInPage: boolean
    }
    className?: string
}

const actionItemClassName = 'action-items__action d-flex justify-content-center align-items-center text-decoration-none'

/**
 *
 */
export const ActionItemsBar = React.memo<ActionItemsBarProps>(props => {
    const { isOpen, barReference } = props.useActionItemsBar()

    const {
        carouselReference,
        canScrollNegative,
        canScrollPositive,
        onNegativeClicked,
        onPositiveClicked,
    } = useCarousel({ direction: 'topToBottom' })

    const haveExtensionsLoaded = useObservable(
        useMemo(() => haveInitialExtensionsLoaded(props.extensionsController.extHostAPI), [props.extensionsController])
    )

    const [isRedesignEnabled] = useRedesignToggle()

    if (!isOpen) {
        return isRedesignEnabled ? <div className="action-items__bar--collapsed " /> : null
    }

    return (
        <div
            className={classNames(
                'action-items__bar p-0 position-relative d-flex flex-column',
                isRedesignEnabled ? 'mr-2' : 'border-left'
                // RepoRevisionContainer content provides the border after redesign
            )}
            ref={barReference}
        >
            {/* To be clear to users that this isn't an error reported by extensions about e.g. the code they're viewing. */}
            <ErrorBoundary location={props.location} render={error => <span>Component error: {error.message}</span>}>
                <ActionItemsDivider />
                {canScrollNegative && (
                    <button
                        type="button"
                        className="btn btn-link action-items__scroll action-items__list-item p-0 border-0"
                        onClick={onNegativeClicked}
                        tabIndex={-1}
                    >
                        <MenuUpIcon className="icon-inline" />
                    </button>
                )}
                <ActionsContainer
                    menu={ContributableMenu.EditorTitle}
                    returnInactiveMenuItems={true}
                    extensionsController={props.extensionsController}
                    empty={null}
                    location={props.location}
                    platformContext={props.platformContext}
                    telemetryService={props.telemetryService}
                >
                    {items => (
                        <ul className="action-items__list list-unstyled m-0" ref={carouselReference}>
                            {items.map((item, index) => {
                                const hasIconURL = !!item.action.actionItem?.iconURL
                                const className = classNames(
                                    actionItemClassName,
                                    !hasIconURL &&
                                        `action-items__action--no-icon action-items__icon-${(index % 5) + 1} text-sm`
                                )
                                const inactiveClassName = classNames(
                                    'action-items__action--inactive',
                                    !hasIconURL && 'action-items__action--no-icon-inactive'
                                )
                                const listItemClassName = classNames(
                                    'action-items__list-item',
                                    index !== items.length - 1 && 'mb-1'
                                )

                                const dataContent = !hasIconURL ? item.action.category?.slice(0, 1) : undefined

                                return (
                                    <li key={item.action.id} className={listItemClassName}>
                                        <ActionItem
                                            {...props}
                                            {...item}
                                            className={className}
                                            dataContent={dataContent}
                                            variant="actionItem"
                                            iconClassName="action-items__icon"
                                            pressedClassName="action-items__action--pressed"
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
                {canScrollPositive && (
                    <button
                        type="button"
                        className="btn btn-link action-items__scroll action-items__list-item p-0 border-0"
                        onClick={onPositiveClicked}
                        tabIndex={-1}
                    >
                        <MenuDownIcon className="icon-inline" />
                    </button>
                )}
                {haveExtensionsLoaded && <ActionItemsDivider />}
                <ul className="list-unstyled m-0">
                    <li className="action-items__list-item">
                        <Link
                            to="/extensions"
                            className={classNames(
                                actionItemClassName,
                                'action-items__list-item action-items__aux-icon'
                            )}
                            data-tooltip="Add extensions"
                        >
                            <PlusIcon className="icon-inline" />
                        </Link>
                    </li>
                </ul>
            </ErrorBoundary>
        </div>
    )
})

export const ActionItemsToggle: React.FunctionComponent<ActionItemsToggleProps> = ({
    useActionItemsToggle,
    extensionsController,
    className,
}) => {
    const { isOpen, toggle, toggleReference, barInPage } = useActionItemsToggle()

    const haveExtensionsLoaded = useObservable(
        useMemo(() => haveInitialExtensionsLoaded(extensionsController.extHostAPI), [extensionsController])
    )

    const [isRedesignEnabled] = useRedesignToggle()

    return barInPage ? (
        <>
            {isRedesignEnabled && <div className="action-items__divider-vertical" />}
            <li
                data-tooltip={`${isOpen ? 'Close' : 'Open'} extensions panel`}
                className={classNames(className, 'nav-item', isRedesignEnabled ? 'mr-2' : 'border-left')}
                // RepoRevisionContainer content provides the border after redesign
            >
                <div
                    className={classNames(
                        'action-items__toggle-container',
                        isOpen && 'action-items__toggle-container--open'
                    )}
                >
                    <ButtonLink
                        className={classNames(
                            actionItemClassName,
                            'action-items__aux-icon',
                            'action-items__action--toggle'
                        )}
                        onSelect={toggle}
                        buttonLinkRef={toggleReference}
                    >
                        {!haveExtensionsLoaded ? (
                            <LoadingSpinner className="icon-inline" />
                        ) : isOpen ? (
                            <ChevronDoubleUpIcon className="icon-inline" />
                        ) : (
                            <PuzzleOutlineIcon className="icon-inline" />
                        )}
                    </ButtonLink>
                </div>
            </li>
        </>
    ) : null
}

const ActionItemsDivider: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <li className={classNames(className, 'action-items__divider-horizontal position-relative rounded-sm d-flex')} />
)
