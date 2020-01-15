import H from 'history'
import * as React from 'react'
import { Subject, Subscription, combineLatest } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { ContributionScope, Context } from '../api/client/context/context'
import { ContributableMenu } from '../api/protocol'
import { getContributedActionItems } from '../contributions/contributions'
import { ExtensionsControllerProps } from '../extensions/controller'
import { PlatformContextProps } from '../platform/context'
import { TelemetryProps } from '../telemetry/telemetryService'
import { ActionItem, ActionItemAction } from './ActionItem'
import { ActionsState } from './actions'

export interface ActionsProps
    extends ExtensionsControllerProps<'executeCommand' | 'services'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings'> {
    menu: ContributableMenu
    scope?: ContributionScope
    extraContext?: Context<any>
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
     * If set, it is rendered when there are no contributed items for that menu. Use null to render nothing when
     * empty.
     */
    empty?: JSX.Element | null
}

/** Displays the actions in a container, with a wrapper and/or empty element. */
export class ActionsContainer extends React.PureComponent<Props, ActionsState> {
    public state: ActionsState = {}

    private scopeChanges = new Subject<ContributionScope | undefined>()
    private extraContextChanges = new Subject<Context<any> | undefined>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        that.subscriptions.add(
            combineLatest([that.scopeChanges, that.extraContextChanges])
                .pipe(
                    switchMap(([scope, extraContext]) =>
                        that.props.extensionsController.services.contribution.getContributions(scope, extraContext)
                    )
                )
                .subscribe(contributions => that.setState({ contributions }))
        )
        that.scopeChanges.next(that.props.scope)
    }

    public componentDidUpdate(prevProps: Props): void {
        if (prevProps.scope !== that.props.scope) {
            that.scopeChanges.next(that.props.scope)
        }
        if (prevProps.extraContext !== that.props.extraContext) {
            that.extraContextChanges.next(that.props.extraContext)
        }
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!that.state.contributions) {
            return null // loading
        }

        const items = getContributedActionItems(that.state.contributions, that.props.menu)
        if (that.props.empty !== undefined && items.length === 0) {
            return that.props.empty
        }

        const render = that.props.render || that.defaultRenderItems
        return render(items)
    }

    private defaultRenderItems = (items: ActionItemAction[]): JSX.Element | null => (
        <>
            {items.map((item, i) => (
                <ActionItem {...that.props} key={i} {...item} />
            ))}
        </>
    )
}
