import React, { useCallback, useMemo } from 'react'
import { LocalStorageSubject } from '../../../../shared/src/util/LocalStorageSubject'
import { useObservable } from '../../../../shared/src/util/useObservable'
import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'
import ChevronDoubleUpIcon from 'mdi-react/ChevronDoubleUpIcon'
import { ButtonLink } from '../../../../shared/src/components/LinkOrButton'
import classNames from 'classnames'
import { ActionsContainer } from '../../../../shared/src/actions/ActionsContainer'
import { ContributableMenu } from '../../../../shared/src/api/protocol'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import * as H from 'history'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../shared/src/telemetry/telemetryService'
import { ActionItem, ActionItemAction } from '../../../../shared/src/actions/ActionItem'
import PlusIcon from 'mdi-react/PlusIcon'
import { Link } from 'react-router-dom'
import { merge, combineLatest, EMPTY, fromEvent, ReplaySubject } from 'rxjs'
import { filter, mapTo, switchMap, tap } from 'rxjs/operators'
import { Key } from 'ts-key-enum'
import { tabbable } from 'tabbable'
import { head, last } from 'lodash'
import { useCarousel } from '../../components/useCarousel'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import MenuDownIcon from 'mdi-react/MenuDownIcon'

// Action items bar and toggle are two separate components due to their placement in the DOM tree

export function useWebActionItems(): Pick<ActionItemsBarProps, 'useActionItemsBar'> &
    Pick<ActionItemsToggleProps, 'useActionItemsToggle'> {
    // Need to pass in contribution point, template type. pass in default open state (we want to keep it closed on search pages by default?)
    // Should toggle state depend on context? or should all action items bars share state for consistency during navigation?
    // use template type dependent on menu/context
    const toggles = useMemo(() => new LocalStorageSubject('action-items-bar-expanded', true), [])

    const toggleReferences = useMemo(() => new ReplaySubject<HTMLElement>(1), [])
    const nextToggleReference = useCallback((toggle: HTMLElement) => toggleReferences.next(toggle), [toggleReferences])

    const barReferences = useMemo(() => new ReplaySubject<HTMLElement>(1), [])
    const nextBarReference = useCallback((bar: HTMLElement) => barReferences.next(bar), [barReferences])

    // Set up keyboard navigation for distant toggle and bar. Removes previous event
    // listeners whenever references change.
    useObservable(
        useMemo(
            () =>
                combineLatest([barReferences, toggleReferences]).pipe(
                    switchMap(([bar, toggle]) => {
                        if (toggle && bar) {
                            const toggleTabs = fromEvent<React.KeyboardEvent>(toggle, 'keydown').pipe(
                                filter(event => event.key === Key.Tab && !event.shiftKey),
                                tap(event => {
                                    const firstBarTabbable = head(tabbable(bar))
                                    if (firstBarTabbable) {
                                        firstBarTabbable.focus()
                                        event.preventDefault()
                                    }
                                })
                            )

                            const barTabs = fromEvent<React.KeyboardEvent>(bar, 'keydown').pipe(
                                filter(event => event.key === Key.Tab),
                                filter(event =>
                                    event.shiftKey ? isFirstTabbable(event.target) : isLastTabbable(event.target)
                                ),
                                tap(event => {
                                    toggle.focus()
                                    event.preventDefault()
                                })
                            )

                            function isLastTabbable(eventTarget: EventTarget): boolean {
                                return eventTarget === last(tabbable(bar))
                            }

                            function isFirstTabbable(eventTarget: EventTarget): boolean {
                                return eventTarget === head(tabbable(bar))
                            }

                            return merge(toggleTabs, barTabs).pipe(
                                // We don't want to rerender the subtree on keydown events
                                mapTo(undefined)
                            )
                        }
                        // Action items bar is not open, don't add event listeners
                        return EMPTY
                    })
                ),
            [barReferences, toggleReferences]
        )
    )

    const useActionItemsBar = useCallback(() => {
        // `useActionItemsBar` will be used as a hook
        // eslint-disable-next-line react-hooks/rules-of-hooks
        const isOpen = useObservable(toggles)

        return { isOpen, barReference: nextBarReference }
    }, [toggles, nextBarReference])

    const useActionItemsToggle = useCallback(() => {
        // `useActionItemsToggle` will be used as a hook
        // eslint-disable-next-line react-hooks/rules-of-hooks
        const isOpen = useObservable(toggles)

        // eslint-disable-next-line react-hooks/rules-of-hooks
        const toggle = useCallback(() => toggles.next(!isOpen), [isOpen])

        return { isOpen, toggle, toggleReference: nextToggleReference }
    }, [toggles, nextToggleReference])

    return {
        useActionItemsBar,
        useActionItemsToggle,
    }
}

export interface ActionItemsBarProps extends ExtensionsControllerProps, PlatformContextProps, TelemetryProps {
    useActionItemsBar: () => { isOpen: boolean | undefined; barReference: React.RefCallback<HTMLElement> }
    location: H.Location
}

export interface ActionItemsToggleProps {
    useActionItemsToggle: () => {
        isOpen: boolean | undefined
        toggle: () => void
        toggleReference: React.RefCallback<HTMLElement>
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

    if (!isOpen) {
        return null
    }

    return (
        <div className="action-items__bar p-0 border-left position-relative d-flex flex-column" ref={barReference}>
            <ActionItemsDivider />
            {canScrollNegative && (
                <button
                    type="button"
                    className="btn btn-link action-items__scroll action-items__list-item p-0 border-0"
                    onClick={onNegativeClicked}
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
                        {[
                            ...items,
                            // TODO(tj): Temporary: testing default icons
                            ...new Array(20).fill(null).map<ActionItemAction>((_value, index) => ({
                                active: true,
                                action: {
                                    category: String(index).slice(-1),
                                    command: 'open',
                                    actionItem: {},
                                    id: `fake-${index}`,
                                },
                            })),
                        ].map((item, index) => (
                            <li key={item.action.id} className="action-items__list-item">
                                <ActionItem
                                    {...props}
                                    {...item}
                                    className={classNames(
                                        actionItemClassName,
                                        !item.action.actionItem?.iconURL &&
                                            `action-items__action--no-icon action-items__icon-${
                                                (index % 5) + 1
                                            } text-sm`
                                    )}
                                    dataContent={
                                        !item.action.actionItem?.iconURL ? item.action.category?.slice(0, 1) : undefined
                                    }
                                    variant="actionItem"
                                    iconClassName="action-items__icon"
                                    pressedClassName="action-items__action--pressed"
                                    inactiveClassName="action-items__action--inactive"
                                    hideLabel={true}
                                />
                            </li>
                        ))}
                    </ul>
                )}
            </ActionsContainer>
            {canScrollPositive && (
                <button
                    type="button"
                    className="btn btn-link action-items__scroll action-items__list-item p-0 border-0"
                    onClick={onPositiveClicked}
                >
                    <MenuDownIcon className="icon-inline" />
                </button>
            )}
            <ActionItemsDivider />
            <ul className="list-unstyled m-0">
                <li className="action-items__list-item">
                    <Link
                        to="/extensions"
                        className={classNames(actionItemClassName, 'action-items__list-item')}
                        data-tooltip="Add extensions"
                    >
                        <PlusIcon className="icon-inline" />
                    </Link>
                </li>
            </ul>
        </div>
    )
})

export const ActionItemsToggle: React.FunctionComponent<ActionItemsToggleProps> = ({
    useActionItemsToggle,
    className,
}) => {
    const { isOpen, toggle, toggleReference } = useActionItemsToggle()

    return (
        <li
            data-tooltip={`${isOpen ? 'Close' : 'Open'} extensions panel`}
            className={classNames(className, 'nav-item border-left')}
        >
            <div
                className={classNames(
                    'action-items__toggle-container',
                    isOpen && 'action-items__toggle-container--open'
                )}
            >
                <ButtonLink
                    className={classNames(actionItemClassName)}
                    onSelect={toggle}
                    buttonLinkRef={toggleReference}
                >
                    {isOpen ? (
                        <ChevronDoubleUpIcon className="icon-inline" />
                    ) : (
                        <PuzzleOutlineIcon className="icon-inline" />
                    )}
                </ButtonLink>
            </div>
        </li>
    )
}

const ActionItemsDivider: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <li className={classNames(className, 'action-items__divider position-relative rounded-sm d-flex')} />
)
