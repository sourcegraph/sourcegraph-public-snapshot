import * as React from 'react'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from '../auth'
import { UserAvatar } from '../settings/user/UserAvatar'

interface Props {}

interface State {
    savedQueries: boolean
    user: GQL.IUser | ImmutableUser | null
}

export class NavLinks extends React.Component<Props, State> {
    private subscriptions = new Subscription()

    constructor() {
        super()
        this.state = {
            savedQueries: false,
            user: window.context.user,
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            currentUser.subscribe(user => {
                this.setState({
                    user: user || window.context.user,
                    savedQueries:
                        (!!user && user.tags && user.tags.some(tag => tag.name === 'saved-queries')) ||
                        (!!user &&
                            user.orgs &&
                            user.orgs.some(org => org.tags && org.tags.some(tag => tag.name === 'saved-queries'))),
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
                {this.state.savedQueries && (
                    <Link to="/search/queries" className="nav-links__link">
                        Queries
                    </Link>
                )}
                {!window.context.onPrem && (
                    <a href="https://about.sourcegraph.com" className="nav-links__link">
                        About
                    </a>
                )}
                {window.context.onPrem && (
                    <Link to="/browse" className="nav-links__link">
                        Browse
                    </Link>
                )}
                {this.state.user ? (
                    <Link className="nav-links__link" to="/settings">
                        <UserAvatar size={64} />
                    </Link>
                ) : (
                    !window.context.onPrem && (
                        <Link className="nav-links__link btn btn-primary" to="/sign-in">
                            Sign in
                        </Link>
                    )
                )}
            </div>
        )
    }
}
