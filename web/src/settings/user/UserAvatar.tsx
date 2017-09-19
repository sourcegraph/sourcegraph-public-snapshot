import { UserWomanAlternate } from '@sourcegraph/icons/lib/UserWomanAlternate'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from '../../auth'

interface Props {
    linkUrl?: string
}

interface State {
    user: GQL.IUser | null
}

export class UserAvatar extends React.Component<Props, State> {
    private subscriptions = new Subscription()

    constructor() {
        super()
        this.state = {
            user: null
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            currentUser.subscribe(
                user => this.setState({ user }),
                error => console.error(error)
            )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const avatar = this.state.user && this.state.user.avatarURL ?
            <img className='avatar-icon' src={this.state.user.avatarURL} /> :
            <UserWomanAlternate />

        if (this.props.linkUrl) {
            return (
                <Link to={this.props.linkUrl} className='avatar'>{avatar}</Link>
            )
        }
        return (
            <div className='avatar'>{avatar}</div>
        )
    }
}
