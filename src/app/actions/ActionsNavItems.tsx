import * as React from 'react'
import { Subscription } from 'rxjs'
import { ConfigurationCascade, ConfigurationSubject } from '../../settings'
import { ActionItem } from './ActionItem'
import { ActionsProps, ActionsState } from './actions'
import { getContributedActionItems } from './contributions'

/**
 * Renders the actions a fragment of <li class="nav-item"> elements, for use in a Bootstrap <ul class="nav"> or <ul
 * class="navbar-nav">.
 */
export class ActionsNavItems<
    S extends ConfigurationSubject,
    C extends ConfigurationCascade<S>
> extends React.PureComponent<ActionsProps<S, C>, ActionsState> {
    public state: ActionsState = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.props.cxpController.registries.contribution.contributions.subscribe(contributions =>
                this.setState({ contributions })
            )
        )
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
                    <li key={i} className="nav-item">
                        <ActionItem
                            key={i}
                            {...item}
                            variant="actionItem"
                            cxpController={this.props.cxpController}
                            extensions={this.props.extensions}
                        />
                    </li>
                ))}
            </>
        )
    }
}
