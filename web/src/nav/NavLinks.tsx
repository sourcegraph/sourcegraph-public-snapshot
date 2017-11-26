import * as React from 'react'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from '../auth'
import { UserAvatar } from '../settings/user/UserAvatar'

interface Props {}

interface State {
    user: GQL.IUser | ImmutableUser | null
}

export class NavLinks extends React.Component<Props, State> {
    private subscriptions = new Subscription()

    constructor() {
        super()
        this.state = {
            user: window.context.user,
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            currentUser.subscribe(user => {
                this.setState({
                    user: user || window.context.user,
                })
            })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="nav-links">
                {!window.context.onPrem && (
                    <a href="https://about.sourcegraph.com" className="nav-links__link">
                        About
                    </a>
                )}
                {// if on-prem, never show a user avatar or sign-in button
                window.context.onPrem ? null : this.state.user ? (
                    <Link className="nav-links__link" to="/settings">
                        <UserAvatar size={64} />
                    </Link>
                ) : (
                    <Link className="nav-links__link btn btn-primary" to="/sign-in">
                        Sign in
                    </Link>
                )}
            </div>
        )
    }
}
