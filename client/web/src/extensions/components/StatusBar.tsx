import React, { useCallback, useMemo, useState } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import { Observable, timer } from 'rxjs'
import { filter, first, mapTo, switchMap } from 'rxjs/operators'

import { urlForClientCommandOpen } from '@sourcegraph/shared/src/actions/ActionItem'
import { StatusBarItemWithKey } from '@sourcegraph/shared/src/api/extension/api/codeEditor'
import { haveInitialExtensionsLoaded } from '@sourcegraph/shared/src/api/features'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { Badge, Button, useObservable, Link, ButtonLink, Icon } from '@sourcegraph/wildcard'

import { ErrorBoundary } from '../../components/ErrorBoundary'
import { useCarousel } from '../../components/useCarousel'

import styles from './StatusBar.module.scss'

interface StatusBarProps extends ExtensionsControllerProps<'extHostAPI' | 'executeCommand'> {
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
}) => {
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

    const {
        carouselReference,
        canScrollNegative,
        canScrollPositive,
        onNegativeClicked,
        onPositiveClicked,
    } = useCarousel({ direction: 'leftToRight' })

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
            ref={statusBarRef}
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
                    >
                        <Icon as={ChevronLeftIcon} />
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
                    >
                        <Icon as={ChevronRightIcon} />
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
        } & ExtensionsControllerProps<'extHostAPI' | 'executeCommand'>
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
        <ButtonLink
            className={classNames(
                'h-100 d-flex align-items-center px-1',
                styles.item,
                noop && classNames('text-decoration-none', styles.itemNoop),
                className
            )}
            data-tooltip={statusBarItem.tooltip}
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
    )
}
