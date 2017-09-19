import * as React from 'react'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from '../../auth'
import { PageTitle } from '../../components/PageTitle'
import { events } from '../../tracking/events'
import { UserAvatar } from './UserAvatar'

interface Props { }
interface State {
    user: GQL.IUser | null
    error?: Error
}

/**
 * A landing page for the user to sign in or register, if not authed
 */
export class UserProfilePage extends React.Component<Props, State> {
    private subscriptions = new Subscription()

    constructor() {
        super()
        this.state = {
            user: null
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(currentUser.subscribe(
            user => this.setState({ user }),
            error => this.setState({ error })
        ))
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className='ui-section'>
                <PageTitle title='profile' />
                <h1>Your Sourcegraph profile</h1>
                <div className='user-profile-page__split-row'>
                    <div className='user-profile-page__avatar-column'>
                        <UserAvatar size={64} />
                    </div>
                    <form className='settings-page__form'>
                        <input readOnly type='text' className='ui-text-box'
                            value={this.state.user && this.state.user.email || ''} placeholder='Email' />
                        {/* TODO(dan): make this form editable
                        <p>
                            <input type='submit' className='settings-ui-button ui-button--right' value='Save' />
                        </p> */}
                    </form>
                </div>
                <div className='user-profile-page__button-row'>
                    <Link to='/editor-auth' className='ui-button user-profile-page__button-spaced'>
                        Authenticate your Sourcegraph editor
                    </Link>
                    <a href='/-/sign-out' onClick={this.logTelemetryOnSignOut} className='ui-button user-profile-page__button-spaced'>
                        Sign out
                    </a>
                </div>
            </div>
        )
    }

    private logTelemetryOnSignOut(): void {
        events.SignOutClicked.log()
    }
}
