import * as H from 'history'
import * as React from 'react'
import { ExtensionsChangeProps, ExtensionsProps } from '../backend/features'
import * as GQL from '../backend/graphqlschema'
import { CXPStatusPopover } from '../cxp/components/CXPStatus'
import { CXPControllerProps, CXPEnvironmentProps, USE_CXP } from '../cxp/CXPEnvironment'
import { ExtensionSelector } from '../registry/extensions/actions/ExtensionSelector'
import { platformEnabled } from '../user/tags'

/** The global debug toolbar is shown when localStorage.debug is truthy. */
const SHOW_GLOBAL_DEBUG = localStorage.getItem('debug') !== null

interface Props extends ExtensionsProps, ExtensionsChangeProps, CXPEnvironmentProps, CXPControllerProps {
    user: GQL.IUser | null
    location: H.Location
    history: H.History
}

/** A global debug toolbar shown in the bottom right of the window. */
export const GlobalDebug: React.SFC<Props> = props =>
    SHOW_GLOBAL_DEBUG ? (
        <div className="global-debug navbar navbar-expand">
            <ul className="navbar-nav align-items-center">
                <li className="nav-item">
                    {props.user &&
                        platformEnabled(props.user) && (
                            <ExtensionSelector
                                key="extension-selector"
                                className="mr-1"
                                onChange={props.onExtensionsChange}
                                configuredExtensionsURL={
                                    (props.user && props.user.configuredExtensions.url) || undefined
                                }
                                history={props.history}
                                location={props.location}
                            />
                        )}
                </li>
                <li className="nav-item">
                    {props.user &&
                        platformEnabled(props.user) &&
                        USE_CXP && (
                            <CXPStatusPopover
                                cxpEnvironment={props.cxpEnvironment}
                                cxpController={props.cxpController}
                            />
                        )}
                </li>
            </ul>
        </div>
    ) : null
