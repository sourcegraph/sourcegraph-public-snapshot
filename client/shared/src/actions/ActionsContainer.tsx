import * as H from 'history'
import React, { useMemo } from 'react'
import { from } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { wrapRemoteObservable } from '../api/client/api/common'
import { ContributionScope, Context } from '../api/client/context/context'
import { ContributableMenu } from '../api/protocol'
import { getContributedActionItems } from '../contributions/contributions'
import { ExtensionsControllerProps } from '../extensions/controller'
import { PlatformContextProps } from '../platform/context'
import { TelemetryProps } from '../telemetry/telemetryService'
import { useObservable } from '../util/useObservable'
import { ActionItem, ActionItemAction } from './ActionItem'

export interface ActionsProps
    extends ExtensionsControllerProps<'executeCommand' | 'extHostAPI'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings'> {
    menu: ContributableMenu
    scope?: ContributionScope
    extraContext?: Context
    listClass?: string
    location: H.Location
}
interface Props extends ActionsProps, TelemetryProps {
    /**
     * Called with the array of contributed items to produce the rendered component. If not set, uses a default
     * render function that renders a <ActionItem> for each item.
     */
    render?: (items: ActionItemAction[]) => JSX.Element | null

    /**
     * If set, it is rendered when there are no contributed items for this menu. Use null to render nothing when
     * empty.
     */
    empty?: JSX.Element | null
}

/** Displays the actions in a container, with a wrapper and/or empty element. */
export const ActionsContainer: React.FunctionComponent<Props> = props => {
    const { scope, extraContext, extensionsController, menu, empty } = props

    const contributions = useObservable(
        useMemo(
            () =>
                from(extensionsController.extHostAPI).pipe(
                    switchMap(extensionHostAPI =>
                        wrapRemoteObservable(extensionHostAPI.getContributions(scope, extraContext))
                    )
                ),
            [scope, extraContext, extensionsController.extHostAPI]
        )
    )

    if (!contributions) {
        return null // loading
    }

    const items = getContributedActionItems(contributions, menu)
    if (empty !== undefined && items.length === 0) {
        return empty
    }

    const render = props.render || defaultRenderItems
    return render(items, props)
}

const defaultRenderItems = (items: ActionItemAction[], props: Props): JSX.Element | null => (
    <>
        {items.map(item => (
            <ActionItem {...props} key={item.action.id} {...item} />
        ))}
    </>
)
