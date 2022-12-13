import React, { useCallback, useEffect, useMemo, useState } from 'react'

import { mdiChevronLeft, mdiChevronRight } from '@mdi/js'
import classNames from 'classnames'
import * as H from 'history'
import { Observable, Subscription, timer } from 'rxjs'
import { filter, first, mapTo, switchMap } from 'rxjs/operators'
import { tabbable } from 'tabbable'
import { useMergeRefs } from 'use-callback-ref'

import { isDefined, logger } from '@sourcegraph/common'
import { urlForClientCommandOpen } from '@sourcegraph/shared/src/actions/ActionItem'
import { StatusBarItemWithKey } from '@sourcegraph/shared/src/api/extension/api/codeEditor'
import { haveInitialExtensionsLoaded } from '@sourcegraph/shared/src/api/features'
import { syncRemoteSubscription } from '@sourcegraph/shared/src/api/util'
import { RequiredExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { Badge, Button, useObservable, Link, ButtonLink, Icon, Tooltip } from '@sourcegraph/wildcard'

import { ErrorBoundary } from '../../components/ErrorBoundary'
import { useCarousel } from '../../components/useCarousel'

import styles from './StatusBar.module.scss'

interface StatusBarProps
    extends RequiredExtensionsControllerProps<'extHostAPI' | 'executeCommand' | 'registerCommand'> {
    getStatusBarItems: () => Observable<StatusBarItemWithKey[] | 'loading'>
    className?: string
    statusBarItemClassName?: string
    /**
     * Used to determine when to restart timer to show "Install extensions"
     * message when there are no status bar items. Only necessary when status bar
     * persists beteween files (e.g. for `<Blob>`).
     */
    uri?: string

    location: H.Location

    statusBarRef?: React.Ref<HTMLDivElement>

    /** Whether to hide the status bar while extensions are loading. */
    hideWhileInitializing?: boolean

    /** If specified, this text will be displayed in a badge left of the status bar items */
    badgeText?: string

    /** If true, the status bar will be focusable through the command palette. */
    isBlobPage?: boolean
}

export const StatusBar: React.FunctionComponent<React.PropsWithChildren<StatusBarProps>> = ({
    getStatusBarItems,
    className,
    statusBarItemClassName,
    extensionsController,
    uri,
    location,
    statusBarRef,
    hideWhileInitializing,
    badgeText,
    isBlobPage,
}) => {
    const mergedReferences = useMergeRefs([statusBarRef].filter(isDefined))

    const statusBarItems = useObservable(useMemo(() => getStatusBarItems(), [getStatusBarItems]))

    // Wait a generous amount of time on top of initial extension loading
    // before showing "Install extensions" message to be forgiving of extensions
    // that make slow network requests and forget to indicate loading state.
    const hasEnoughTimePassed = useObservable(
        useMemo(
            () =>
                haveInitialExtensionsLoaded(extensionsController.extHostAPI).pipe(
                    filter(haveLoaded => haveLoaded),
                    first(),
                    switchMap(() => timer(2000).pipe(mapTo(true)))
                ),
            // We want to recreate the observable on uri change, so keep
            // the unnecessary dependency.
            // eslint-disable-next-line react-hooks/exhaustive-deps
            [uri, extensionsController]
        )
    )

    // Add "focus status bar" action to command palette.
    useEffect(() => {
        const subscription = new Subscription()

        // Only when one status bar is rendered (i.e. is blob page).
        if (isBlobPage) {
            extensionsController.extHostAPI
                .then(extensionHostAPI => {
                    subscription.add(
                        syncRemoteSubscription(
                            extensionHostAPI.registerContributions({
                                menus: {
                                    commandPalette: [
                                        {
                                            action: 'focusStatusBar',
                                        },
                                    ],
                                },
                                actions: [
                                    {
                                        id: 'focusStatusBar',
                                        title: 'Focus status bar',
                                        command: 'focusStatusBar',
                                    },
                                ],
                            })
                        )
                    )

                    subscription.add(
                        extensionsController.registerCommand({
                            command: 'focusStatusBar',
                            run: () => {
                                const statusBarElement = mergedReferences.current
                                if (statusBarElement) {
                                    tabbable(statusBarElement)[0]?.focus()
                                }
                                return Promise.resolve()
                            },
                        })
                    )
                })
                .catch(error => logger.error('Error registering "Focus status bar" command', error))
        }

        return () => subscription.unsubscribe()
    }, [extensionsController, isBlobPage, mergedReferences])

    const { carouselReference, canScrollNegative, canScrollPositive, onNegativeClicked, onPositiveClicked } =
        useCarousel({ direction: 'leftToRight' })

    if (!hasEnoughTimePassed && hideWhileInitializing) {
        return null
    }

    return (
        <div
            className={classNames(
                styles.statusBar,
                'border-top',
                'percy-hide', // TODO: Fix flaky status bar in Percy tests: https://github.com/sourcegraph/sourcegraph/issues/20751
                className
            )}
            ref={mergedReferences}
        >
            <ErrorBoundary
                location={location}
                // To be clear to users that this isn't an error reported by extensions
                // about e.g. the code they're viewing.
                render={error => (
                    <div className={classNames('ml-2', styles.item)}>
                        <small className="text-muted">Status bar component error: {error.message}</small>
                    </div>
                )}
            >
                {canScrollNegative && (
                    <Button
                        className={classNames('border-0', styles.scroll)}
                        onClick={onNegativeClicked}
                        variant="link"
                        aria-label="Scroll left"
                    >
                        <Icon aria-hidden={true} svgPath={mdiChevronLeft} />
                    </Button>
                )}
                <div className={classNames('d-flex align-items-center px-2', styles.items)} ref={carouselReference}>
                    {badgeText && (
                        <Badge variant="secondary" className="m-0" as="p">
                            {badgeText}
                        </Badge>
                    )}
                    {!!statusBarItems && statusBarItems !== 'loading' && statusBarItems.length > 0
                        ? statusBarItems.map(statusBarItem => (
                              <StatusBarItem
                                  key={statusBarItem.key}
                                  statusBarItem={statusBarItem}
                                  extensionsController={extensionsController}
                                  location={location}
                                  className={statusBarItemClassName}
                              />
                          ))
                        : hasEnoughTimePassed && (
                              <div className={classNames('ml-2', styles.item)}>
                                  <small className="text-muted">
                                      No information from extensions available.{' '}
                                      <Link to="/extensions">
                                          Find extensions in the Sourcegraph extension registry
                                      </Link>
                                  </small>
                              </div>
                          )}
                </div>
                {canScrollPositive && (
                    <Button
                        className={classNames('border-0', styles.scroll)}
                        onClick={onPositiveClicked}
                        variant="link"
                        aria-label="Scroll right"
                    >
                        <Icon aria-hidden={true} svgPath={mdiChevronRight} />
                    </Button>
                )}
            </ErrorBoundary>
        </div>
    )
}

const StatusBarItem: React.FunctionComponent<
    React.PropsWithChildren<
        {
            statusBarItem: StatusBarItemWithKey
            className?: string
            component?: JSX.Element
            location: H.Location
        } & RequiredExtensionsControllerProps<'extHostAPI' | 'executeCommand'>
    >
> = ({ statusBarItem, className, component, extensionsController, location }) => {
    const [commandState, setCommandState] = useState<'loading' | null>(null)

    const command = useMemo(() => statusBarItem.command, [statusBarItem.command])

    const to = useMemo(
        () =>
            command && urlForClientCommandOpen({ command: command.id, commandArguments: command.args }, location.hash),
        [command, location]
    )

    const handleCommand = useCallback(() => {
        // Do not execute the command if `to` is defined.
        // The <ButtonLink>'s default event handler will do what we want (which is to open a URL).
        if (commandState !== 'loading' && command && !to) {
            setCommandState('loading')
            extensionsController
                .executeCommand({ command: command.id, args: command.args })
                .then(() => {
                    setCommandState(null)
                })
                .catch(() => {
                    // noop, errors will be displayed as notifications
                    setCommandState(null)
                })
        }
    }, [commandState, extensionsController, command, to])

    const noop = !command

    return (
        <Tooltip content={statusBarItem.tooltip}>
            <ButtonLink
                className={classNames(
                    'h-100 d-flex align-items-center px-1',
                    styles.item,
                    noop && classNames('text-decoration-none', styles.itemNoop),
                    className
                )}
                onSelect={handleCommand}
                tabIndex={noop ? -1 : 0}
                to={to}
                disabled={commandState === 'loading'}
            >
                {component || (
                    <small className={classNames(styles.text, commandState === 'loading' && 'text-muted')}>
                        {statusBarItem.text}
                    </small>
                )}
            </ButtonLink>
        </Tooltip>
    )
}
