import React, { useMemo } from 'react'

import type * as H from 'history'
import { from } from 'rxjs'
import { switchMap } from 'rxjs/operators'

import type { ContributableMenu } from '@sourcegraph/client-api'
import { useObservable } from '@sourcegraph/wildcard'

import { wrapRemoteObservable } from '../api/client/api/common'
import type { ContributionOptions } from '../api/extension/extensionHostApi'
import { getContributedActionItems } from '../contributions/contributions'
import type { RequiredExtensionsControllerProps } from '../extensions/controller'
import type { PlatformContextProps } from '../platform/context'
import { TelemetryV2Props } from '../telemetry'
import type { TelemetryProps } from '../telemetry/telemetryService'

import { ActionItem, type ActionItemAction } from './ActionItem'

export interface ActionsProps
    extends RequiredExtensionsControllerProps<'executeCommand' | 'extHostAPI'>,
        PlatformContextProps<'settings'>,
        ContributionOptions {
    menu: ContributableMenu
    listClass?: string
    location: H.Location
}
interface Props extends ActionsProps, TelemetryProps, TelemetryV2Props {
    /**
     * Called with the array of contributed items to produce the rendered component. If not set, uses a default
     * render function that renders a <ActionItem> for each item.
     */
    children?: (items: ActionItemAction[]) => JSX.Element | null

    /**
     * If set, it is rendered when there are no contributed items for this menu. Use null to render nothing when
     * empty.
     */
    empty?: JSX.Element | null
}

/** Displays the actions in a container, with a wrapper and/or empty element. */
export const ActionsContainer: React.FunctionComponent<Props> = props => {
    const { scope, extraContext, returnInactiveMenuItems, extensionsController, menu, empty } = props

    const contributions = useObservable(
        useMemo(
            () =>
                from(extensionsController.extHostAPI).pipe(
                    switchMap(extensionHostAPI =>
                        wrapRemoteObservable(
                            extensionHostAPI.getContributions({ scope, extraContext, returnInactiveMenuItems })
                        )
                    )
                ),
            [scope, extraContext, returnInactiveMenuItems, extensionsController.extHostAPI]
        )
    )

    if (!contributions) {
        return null // loading
    }

    const items = getContributedActionItems(contributions, menu)
    if (empty !== undefined && items.length === 0) {
        return empty
    }

    const render = props.children || defaultRenderItems
    return render(items, props)
}

const defaultRenderItems = (items: ActionItemAction[], props: Props): JSX.Element | null => (
    <>
        {items.map(item => (
            <ActionItem {...props} key={item.action.id} {...item} />
        ))}
    </>
)
