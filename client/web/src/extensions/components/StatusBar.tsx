import React, { useMemo, useState } from 'react'
import { StatusBarItemWithKey } from '../../../../shared/src/api/extension/api/codeEditor'
import { useObservable } from '../../../../shared/src/util/useObservable'
import classNames from 'classnames'
import { Observable } from 'rxjs'
import { useCarousel } from '../../components/useCarousel'
import MenuLeftIcon from 'mdi-react/MenuLeftIcon'
import MenuRightIcon from 'mdi-react/MenuRightIcon'
import { Link, useLocation } from 'react-router-dom'
import { haveInitialExtensionsLoaded } from '../../../../shared/src/api/features'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { ButtonLink } from '../../../../shared/src/components/LinkOrButton'
import { urlForClientCommandOpen } from '../../../../shared/src/actions/ActionItem'
import { ErrorBoundary } from '../../components/ErrorBoundary'
import * as H from 'history'

interface StatusBarProps extends ExtensionsControllerProps {
    getStatusBarItems: () => Observable<StatusBarItemWithKey[] | 'loading'>
    className?: string
}

export const StatusBar: React.FunctionComponent<StatusBarProps> = ({
    getStatusBarItems,
    className,
    extensionsController,
}) => {
    const statusBarItems = useObservable(useMemo(() => getStatusBarItems(), [getStatusBarItems]))

    const haveExtensionsLoaded = useObservable(
        useMemo(() => haveInitialExtensionsLoaded(extensionsController.extHostAPI), [extensionsController])
    )

    // We need timer based on uri!

    const {
        carouselReference,
        canScrollNegative,
        canScrollPositive,
        onNegativeClicked,
        onPositiveClicked,
    } = useCarousel({ direction: 'leftToRight' })

    const location = useLocation()

    return (
        <div className={classNames('status-bar w-100 border-top d-flex', className)}>
            <ErrorBoundary
                location={location}
                // To be clear to users that this isn't an error reported by extensions
                // about e.g. the code they're viewing.
                render={error => (
                    <div className={`${className}__item ml-2`}>
                        <small className="text-muted">Status bar component error: {error.message}</small>
                    </div>
                )}
            >
                {canScrollNegative && (
                    <button
                        type="button"
                        className="btn btn-link status-bar__scroll border-0"
                        onClick={onNegativeClicked}
                    >
                        <MenuLeftIcon className="icon-inline" />
                    </button>
                )}
                <div className="status-bar__items d-flex px-2" ref={carouselReference}>
                    {!!statusBarItems && statusBarItems !== 'loading' && statusBarItems.length > 0
                        ? statusBarItems.map(statusBarItem => (
                              <StatusBarItem
                                  key={statusBarItem.key}
                                  statusBarItem={statusBarItem}
                                  extensionsController={extensionsController}
                                  location={location}
                              />
                          ))
                        : haveExtensionsLoaded && (
                              <StatusBarItem
                                  key="none-found"
                                  statusBarItem={{
                                      key: 'none-found',
                                      text:
                                          'No information from extensions available. Find extensions in the Sourcegraph extension registry',
                                  }}
                                  extensionsController={extensionsController}
                                  component={
                                      <small className="text-muted">
                                          No information from extensions available.{' '}
                                          <Link to="/extensions">
                                              Find extensions in the Sourcegraph extension registry
                                          </Link>
                                      </small>
                                  }
                                  location={location}
                              />
                          )}
                </div>
                {canScrollPositive && (
                    <button
                        type="button"
                        className="btn btn-link status-bar__scroll border-0"
                        onClick={onPositiveClicked}
                    >
                        <MenuRightIcon className="icon-inline" />
                    </button>
                )}
            </ErrorBoundary>
        </div>
    )
}

const StatusBarItem: React.FunctionComponent<
    {
        statusBarItem: StatusBarItemWithKey
        className?: string
        component?: JSX.Element
        location: H.Location
    } & ExtensionsControllerProps
> = ({ statusBarItem, className = 'status-bar', component, extensionsController, location }) => {
    const [commandState, setCommandState] = useState<'loading' | null>(null)

    const to =
        statusBarItem.command &&
        urlForClientCommandOpen(
            { command: statusBarItem.command.id, commandArguments: statusBarItem.command.args },
            location
        )

    const handleCommand = () => {
        // Do not execute the command if `to` is defined.
        // The <ButtonLink>'s default event handler will do what we want (which is to open a URL).
        if (commandState !== 'loading' && statusBarItem.command && !to) {
            setCommandState('loading')
            extensionsController
                .executeCommand({ command: statusBarItem.command.id, args: statusBarItem.command.args })
                .then(() => {
                    setCommandState(null)
                })
                .catch(error => {
                    // noop, errors will be displayed as notifications
                    console.log('error??', error)
                    setCommandState(null)
                })
        }
    }

    const noop = !statusBarItem.command

    return (
        <ButtonLink
            className={classNames(
                `${className}__item h-100 d-flex align-items-center px-1 text-decoration-none`,
                statusBarItem.tooltip && `${className}__item--tooltipped`,
                noop && `${className}__item--noop`
            )}
            data-tooltip={statusBarItem.tooltip}
            onSelect={handleCommand}
            tabIndex={noop ? -1 : 0}
            to={to}
            disabled={commandState === 'loading'}
        >
            {component || (
                <small className={classNames(`${className}__text`, commandState === 'loading' && 'text-muted')}>
                    {statusBarItem.text}
                </small>
            )}
        </ButtonLink>
    )
}
