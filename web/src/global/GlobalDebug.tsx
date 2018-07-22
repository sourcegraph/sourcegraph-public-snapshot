import * as H from 'history'
import * as React from 'react'
import { ExtensionsChangeProps, ExtensionsProps } from '../backend/features'
import * as GQL from '../backend/graphqlschema'
import { CXPStatusPopover } from '../cxp/components/CXPStatus'
import { CXPControllerProps, CXPEnvironmentProps } from '../cxp/CXPEnvironment'
import { ExtensionSelector } from '../registry/extensions/actions/ExtensionSelector'
import { platformEnabled } from '../user/tags'

interface Props extends ExtensionsProps, ExtensionsChangeProps, CXPEnvironmentProps, CXPControllerProps {
    user: GQL.IUser | null
    location: H.Location
    history: H.History
}

/**
 * A global debug toolbar shown in the bottom right of the window.
 *
 * It is only useful for platform debug, so it's only shown for platform-enabled users.
 */
export const GlobalDebug: React.SFC<Props> = props =>
    platformEnabled(props.user) ? (
        <div className="global-debug navbar navbar-expand">
            <ul className="navbar-nav align-items-center">
                <li className="nav-item">
                    <ExtensionSelector
                        key="extension-selector"
                        className="mr-1"
                        onChange={props.onExtensionsChange}
                        configuredExtensionsURL={(props.user && props.user.configuredExtensions.url) || undefined}
                        history={props.history}
                        location={props.location}
                    />
                </li>
                <li className="nav-item">
                    <CXPStatusPopover cxpEnvironment={props.cxpEnvironment} cxpController={props.cxpController} />
                </li>
            </ul>
        </div>
    ) : null
