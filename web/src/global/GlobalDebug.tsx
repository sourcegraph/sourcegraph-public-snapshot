import * as React from 'react'
import * as GQL from '../backend/graphqlschema'
import { CXPStatusPopover } from '../cxp/components/CXPStatus'
import { CXPControllerProps, CXPEnvironmentProps } from '../cxp/CXPEnvironment'
import { platformEnabled } from '../user/tags'

interface Props extends CXPEnvironmentProps, CXPControllerProps {
    user: GQL.IUser | null
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
                    <CXPStatusPopover cxpEnvironment={props.cxpEnvironment} cxpController={props.cxpController} />
                </li>
            </ul>
        </div>
    ) : null
