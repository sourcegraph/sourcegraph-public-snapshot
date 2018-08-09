import { CXPControllerProps } from '@sourcegraph/extensions-client-common/lib/cxp/controller'
import { CXPStatusPopover } from '@sourcegraph/extensions-client-common/lib/cxp/CXPStatus'
import { CaretDown } from '@sourcegraph/icons/lib/CaretDown'
import { Loader } from '@sourcegraph/icons/lib/Loader'
import * as React from 'react'
import * as GQL from '../backend/graphqlschema'
import { CXPEnvironmentProps } from '../cxp/CXPEnvironment'
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
                    <CXPStatusPopover
                        loaderIcon={Loader as React.ComponentType<{ className: string; onClick?: () => void }>}
                        caretIcon={CaretDown as React.ComponentType<{ className: string; onClick?: () => void }>}
                        cxpController={props.cxpController}
                    />
                </li>
            </ul>
        </div>
    ) : null
