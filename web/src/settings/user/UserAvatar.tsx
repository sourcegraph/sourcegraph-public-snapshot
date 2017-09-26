import { UserWomanAlternate } from '@sourcegraph/icons/lib/UserWomanAlternate'
import * as React from 'react'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from '../../auth'

export interface Avatarable {
    avatarURL: string | null
}

interface Props {
    onClick?: () => void
    size?: number
    user?: Avatarable
}

interface State {
    user: Avatarable | null
}

/**
 * UserAvatar displays the avatar of an Avatarable object
 */
export class UserAvatar extends React.Component<Props, State> {
    private subscriptions = new Subscription()

    constructor() {
        super()
        this.state = {
            user: null
        }
    }

    public componentDidMount(): void {
        if (this.props.user) {
            this.setState({ user: this.props.user })
        } else {
            this.subscriptions.add(
                currentUser.subscribe(
                    user => this.setState({ user }),
                    error => console.error(error)
                )
            )
        }
    }

    public componentWillReceiveProps(nextProps: Props): void {
        if (nextProps.user) {
            this.setState({ user: nextProps.user })
        }
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

        return (
            <div onClick={this.props.onClick} className='avatar'>{avatar}</div>
        )
    }
}
