
import * as React from 'react'
import { Link } from 'react-router-dom'
import { SearchBox } from '../search/SearchBox'
import { SignOutButton } from '../settings/auth/SignOutButton'
import { UserAvatar } from '../settings/user/UserAvatar'
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
                            <Link to='/sign-in' className='ui-button'>Sign in</Link>
                    }
                    {
                        this.state.showSignOut && <SignOutButton />
                    }
                </div>
            </div>
        )
    }
}
