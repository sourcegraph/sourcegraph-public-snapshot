import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { ExtensionStatusPopover } from '../../../shared/src/extensions/ExtensionStatus'
import { ExtensionsEnvironmentProps } from '../extensions/environment/ExtensionsEnvironment'

interface Props extends ExtensionsEnvironmentProps, ExtensionsControllerProps {
    location: H.Location
}

const SHOW_DEBUG = localStorage.getItem('debug') !== null

const ExtensionLink: React.FunctionComponent<{ id: string }> = props => (
    <Link to={`extensions/${props.id}`}>{props.id}</Link>
)

/**
 * A global debug toolbar shown in the bottom right of the window.
 */
export const GlobalDebug: React.FunctionComponent<Props> = props =>
    SHOW_DEBUG ? (
        <div className="global-debug navbar navbar-expand">
            <ul className="navbar-nav align-items-center">
                <li className="nav-item">
                    <ExtensionStatusPopover
                        link={ExtensionLink}
                        location={props.location}
                        extensionsController={props.extensionsController}
                    />
                </li>
            </ul>
        </div>
    ) : null
