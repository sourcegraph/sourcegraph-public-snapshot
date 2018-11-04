import { ExtensionStatusPopover } from '@sourcegraph/extensions-client-common/lib/app/ExtensionStatus'
import { Controller as ClientController } from '@sourcegraph/extensions-client-common/lib/client/controller'
import { ConfigurationSubject, Settings } from '@sourcegraph/extensions-client-common/lib/settings'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import * as React from 'react'
import { sourcegraphUrl } from '../util/context'
import { ShortcutProvider } from './ShortcutProvider'

interface Props {
    location: H.Location
    extensionsController: ClientController<ConfigurationSubject, Settings>
}

const SHOW_DEBUG = localStorage.getItem('debug') !== null

const ExtensionLink: React.SFC<{ id: string }> = props => {
    const extensionURL = new URL(sourcegraphUrl)
    extensionURL.pathname = `extensions/${props.id}`
    return <a href={extensionURL.href}>{props.id}</a>
}

/**
 * A global debug toolbar shown in the bottom right of the window.
 */
export const GlobalDebug: React.SFC<Props> = props =>
    SHOW_DEBUG ? (
        <div className="global-debug navbar navbar-expand">
            <ul className="navbar-nav align-items-center">
                <li className="nav-item">
                    <ShortcutProvider>
                        <ExtensionStatusPopover
                            location={props.location}
                            loaderIcon={
                                LoadingSpinner as React.ComponentType<{ className: string; onClick?: () => void }>
                            }
                            caretIcon={MenuDownIcon as React.ComponentType<{ className: string; onClick?: () => void }>}
                            extensionsController={props.extensionsController}
                            link={ExtensionLink}
                        />
                    </ShortcutProvider>
                </li>
            </ul>
        </div>
    ) : null
