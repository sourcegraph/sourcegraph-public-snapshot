
import * as React from 'react'
import { Link } from 'react-router-dom'
import { SearchBox } from '../search/SearchBox'
import { SignInButton } from '../settings/auth/SignInButton'
import { UserAvatar } from '../settings/user/UserAvatar'
import { events } from '../tracking/events'
import { ParsedRouteProps } from '../util/routes'
import { sourcegraphContext } from '../util/sourcegraphContext'

interface State {
    showSignOut: boolean
}

export class Navbar extends React.Component<ParsedRouteProps, State> {
    public state = {
        showSignOut: false
    }

    public render(): JSX.Element | null {
        return (
            <div className='navbar'>
                <div className='navbar__left'>
                    <Link to='/search' className='navbar__logo-link'>
                        <img className='navbar__logo' src='/.assets/img/sourcegraph-mark.svg' />
                    </Link>
                </div>
                <div className='navbar__search-box-container'>
                    <SearchBox {...this.props} />
                </div>
                <div className='navbar__right'>
                    {
                        sourcegraphContext.user ?
                            <UserAvatar size={64} onClick={() => this.setState({ showSignOut: !this.state.showSignOut })} /> :
                            <SignInButton />
                    }
                    {
                        this.state.showSignOut &&
                            <a href='/-/sign-out' onClick={events.SignOutClicked.log} className='ui-button'>
                                Sign out
                            </a>
                    }
                </div>
            </div>
        )
    }
}
