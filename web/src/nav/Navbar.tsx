
import * as React from 'react'
import { Link } from 'react-router-dom'
import { SearchBox } from '../search/SearchBox'
import { SignInButton } from '../settings/auth/SignInButton'
import { UserAvatar } from '../settings/user/UserAvatar'
import { ParsedRouteProps } from '../util/routes'
import { sourcegraphContext } from '../util/sourcegraphContext'

export class Navbar extends React.Component<ParsedRouteProps, {}> {
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
                            <UserAvatar linkUrl='/settings' /> :
                            <SignInButton />
                    }
                </div>
            </div>
        )
    }
}
