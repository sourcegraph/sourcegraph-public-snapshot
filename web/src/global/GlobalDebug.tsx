import { ExtensionStatusPopover } from '@sourcegraph/extensions-client-common/lib/app/ExtensionStatus'
import { CaretDown } from '@sourcegraph/icons/lib/CaretDown'
import { Loader } from '@sourcegraph/icons/lib/Loader'
import * as React from 'react'
import * as GQL from '../backend/graphqlschema'
import { ExtensionsEnvironmentProps, USE_PLATFORM } from '../extensions/environment/ExtensionsEnvironment'
import { ExtensionsControllerProps } from '../extensions/ExtensionsClientCommonContext'

interface Props extends ExtensionsEnvironmentProps, ExtensionsControllerProps {
    user: GQL.IUser | null
}

/**
 * A global debug toolbar shown in the bottom right of the window.
 *
 * It is only useful for platform debug, so it's only shown for platform-enabled users.
 */
export const GlobalDebug: React.SFC<Props> = props =>
    USE_PLATFORM ? (
        <div className="global-debug navbar navbar-expand">
            <ul className="navbar-nav align-items-center">
                <li className="nav-item">
                    <ExtensionStatusPopover
                        loaderIcon={Loader as React.ComponentType<{ className: string; onClick?: () => void }>}
                        caretIcon={CaretDown as React.ComponentType<{ className: string; onClick?: () => void }>}
                        extensionsController={props.extensionsController}
                    />
                </li>
            </ul>
        </div>
    ) : null
