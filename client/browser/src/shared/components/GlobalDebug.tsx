import * as React from 'react'

import * as H from 'history'

import { Controller as ClientController } from '@sourcegraph/shared/src/extensions/controller'
import { ExtensionDevelopmentToolsPopover } from '@sourcegraph/shared/src/extensions/devtools'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Link } from '@sourcegraph/wildcard'

import { ShortcutProvider } from './ShortcutProvider'

interface Props extends PlatformContextProps<'sideloadedExtensionURL' | 'settings'> {
    location: H.Location
    extensionsController: ClientController
    sourcegraphURL: string
}

const makeExtensionLink = (
    sourcegraphURL: string
): React.FunctionComponent<React.PropsWithChildren<{ id: string }>> => props => {
    const extensionURL = new URL(sourcegraphURL)
    extensionURL.pathname = `extensions/${props.id}`
    return <Link to={extensionURL.href}>{props.id}</Link>
}

/**
 * A global debug toolbar shown in the bottom right of the window.
 */
export const GlobalDebug: React.FunctionComponent<React.PropsWithChildren<Props>> = props => (
    <div className="navbar navbar-expand" data-global-debug={true}>
        <div className="navbar-nav align-items-center">
            <div className="nav-item">
                <ShortcutProvider>
                    <ExtensionDevelopmentToolsPopover
                        extensionsController={props.extensionsController}
                        link={makeExtensionLink(props.sourcegraphURL)}
                        platformContext={props.platformContext}
                    />
                </ShortcutProvider>
            </div>
        </div>
    </div>
)
