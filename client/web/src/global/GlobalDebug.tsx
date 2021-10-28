import * as H from 'history'

import { ErrorBoundary } from '../components/ErrorBoundary'
import { ExtensionDevelopmentToolsPopover } from '@sourcegraph/shared/src/extensions/devtools'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import React from 'react'
import classNames from 'classnames'
import styles from './GlobalDebug.module.scss'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    location: H.Location
}

const SHOW_DEBUG = localStorage.getItem('debug') !== null

const ExtensionLink: React.FunctionComponent<{ id: string }> = props => (
    <Link to={`/extensions/${props.id}`}>{props.id}</Link>
)

const ExtensionDevelopmentToolsError = (error: Error): JSX.Element => (
    <span>Error rendering extension development tools: {error.message}</span>
)

/**
 * A global debug toolbar shown in the bottom right of the window.
 */
export const GlobalDebug: React.FunctionComponent<Props> = props =>
    SHOW_DEBUG ? (
        <ul className={classNames('nav', styles.globalDebug)}>
            <li className="nav-item">
                <ErrorBoundary location={props.location} render={ExtensionDevelopmentToolsError}>
                    <ExtensionDevelopmentToolsPopover
                        link={ExtensionLink}
                        extensionsController={props.extensionsController}
                        platformContext={props.platformContext}
                    />
                </ErrorBoundary>
            </li>
        </ul>
    ) : null
