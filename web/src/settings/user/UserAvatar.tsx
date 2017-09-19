import { UserWomanAlternate } from '@sourcegraph/icons/lib/UserWomanAlternate'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from '../../auth'

interface Props {
    linkUrl?: string
    size?: number
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
        let avatar: JSX.Element
        if (this.state.user && this.state.user.avatarURL) {
            const url = new URL(this.state.user.avatarURL)
            if (this.props.size) {
                url.searchParams.set('s', this.props.size + '')
            }
            avatar = <img className='avatar-icon' src={url.href} />
        } else {
            avatar = <UserWomanAlternate />
        }

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
