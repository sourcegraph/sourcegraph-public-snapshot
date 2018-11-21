import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { TextDocumentItem } from '../api/client/types/textDocument'
import { getContributedActionItems } from '../contributions/contributions'
import { ActionItem, ActionItemProps } from './ActionItem'
import { ActionsProps, ActionsState } from './actions'

interface ActionsContainerProps extends ActionsProps {
    /**
     * Called with the array of contributed items to produce the rendered component. If not set, uses a default
     * render function that renders a <ActionItem> for each item.
     */
    render?: (items: ActionItemProps[]) => React.ReactElement<any>

    /**
     * If set, it is rendered when there are no contributed items for this menu. Use null to render nothing when
     * empty.
     */
    empty?: React.ReactElement<any> | null
}

interface ActionsContainerState extends ActionsState {}

/** Displays the actions in a container, with a wrapper and/or empty element. */
export class ActionsContainer extends React.PureComponent<ActionsContainerProps, ActionsContainerState> {
    public state: ActionsState = {}

    private scopeChanges = new Subject<TextDocumentItem | undefined>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.scopeChanges
                .pipe(
                    switchMap(scope => this.props.extensionsController.registries.contribution.getContributions(scope))
                )
                .subscribe(contributions => this.setState({ contributions }))
        )
        this.scopeChanges.next(this.props.scope)
    }

    public componentDidUpdate(prevProps: ActionsContainerProps): void {
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

    private defaultRenderItems = (items: ActionItemProps[]): JSX.Element | null => (
        <>
            {items.map((item, i) => (
                <ActionItem
                    key={i}
                    {...item}
                    extensionsController={this.props.extensionsController}
                    platformContext={this.props.platformContext}
                    location={this.props.location}
                />
            ))}
        </>
    )
}
