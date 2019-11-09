import * as H from 'history'
import * as React from 'react'
import { Controller as ClientController } from '../../../../shared/src/extensions/controller'
import { ExtensionStatusPopover } from '../../../../shared/src/extensions/ExtensionStatus'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { ShortcutProvider } from './ShortcutProvider'

interface Props extends PlatformContextProps<'sideloadedExtensionURL' | 'sourcegraphURL'> {
    location: H.Location
    extensionsController: ClientController
}

const makeExtensionLink = (sourcegraphURL: URL): React.FunctionComponent<{ id: string }> => props => {
    const extensionURL = new URL(`extensions/${props.id}`, sourcegraphURL)
    return <a href={extensionURL.href}>{props.id}</a>
}

/**
 * A global debug toolbar shown in the bottom right of the window.
 */
export const GlobalDebug: React.FunctionComponent<Props> = props => {
    const sourcegraphURL = useObservable(props.platformContext.sourcegraphURL)
    if (!sourcegraphURL) {
        return null
    }
    return (
        <div className="global-debug navbar navbar-expand">
            <div className="navbar-nav align-items-center">
                <div className="nav-item">
                    <ShortcutProvider>
                        <ExtensionStatusPopover
                            extensionsController={props.extensionsController}
                            link={makeExtensionLink(sourcegraphURL)}
                            platformContext={props.platformContext}
                        />
                    </ShortcutProvider>
                </div>
            </div>
        </div>
    )
}
