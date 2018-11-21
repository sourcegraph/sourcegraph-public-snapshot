import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { TextDocumentItem } from '../api/client/types/textDocument'
import { getContributedActionItems } from '../contributions/contributions'
import { ActionItem } from './ActionItem'
import { ActionsProps, ActionsState } from './actions'

/**
 * Renders the actions as a fragment of <li class="nav-item"> elements, for use in a Bootstrap <ul
 * class="nav"> or <ul class="navbar-nav">.
 */
export class ActionsNavItems extends React.PureComponent<ActionsProps, ActionsState> {
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

    public componentDidUpdate(prevProps: ActionsProps): void {
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

        return (
            <>
                {getContributedActionItems(this.state.contributions, this.props.menu).map((item, i) => (
                    <li key={i} className={this.props.listClass || 'nav-item'}>
                        <ActionItem
                            key={i}
                            {...item}
                            variant="actionItem"
                            extensionsController={this.props.extensionsController}
                            platformContext={this.props.platformContext}
                            className={this.props.actionItemClass}
                            location={this.props.location}
                        />
                    </li>
                ))}
            </>
        )
    }
}
