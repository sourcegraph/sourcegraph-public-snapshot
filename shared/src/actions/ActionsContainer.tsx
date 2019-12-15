import H from 'history'
import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { ContributionScope } from '../api/client/context/context'
import { ContributableMenu } from '../api/protocol'
import { getContributedActionItems } from '../contributions/contributions'
import { ExtensionsControllerProps } from '../extensions/controller'
import { PlatformContextProps } from '../platform/context'
import { TelemetryProps } from '../telemetry/telemetryService'
import { ActionItem, ActionItemAction } from './ActionItem'
import { ActionsState } from './actions'

export interface ActionsProps
    extends ExtensionsControllerProps<'executeCommand' | 'services'>,
        PlatformContextProps<'forceUpdateTooltip'> {
    menu: ContributableMenu
    scope?: ContributionScope
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
export class ActionsContainer extends React.PureComponent<Props, ActionsState> {
    public state: ActionsState = {}

    private scopeChanges = new Subject<ContributionScope | undefined>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.scopeChanges
                .pipe(switchMap(scope => this.props.extensionsController.services.contribution.getContributions(scope)))
                .subscribe(contributions => this.setState({ contributions }))
        )
        this.scopeChanges.next(this.props.scope)
    }

    public componentDidUpdate(prevProps: Props): void {
        if (prevProps.scope !== this.props.scope) {
            this.scopeChanges.next(this.props.scope)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.contributions) {
            return null // loading
        }

        const items = getContributedActionItems(this.state.contributions, this.props.menu)
        if (this.props.empty !== undefined && items.length === 0) {
            return this.props.empty
        }

        const render = this.props.render || this.defaultRenderItems
        return render(items)
    }

    private defaultRenderItems = (items: ActionItemAction[]): JSX.Element | null => (
        <>
            {items.map((item, i) => (
                <ActionItem {...this.props} key={i} {...item} />
            ))}
        </>
    )
}
