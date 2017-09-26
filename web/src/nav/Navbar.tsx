
import * as React from 'react'
import { Link } from 'react-router-dom'
import { SearchBox } from '../search/SearchBox'
import { UserAvatar } from '../settings/user/UserAvatar'
import { ParsedRouteProps } from '../util/routes'
import { sourcegraphContext } from '../util/sourcegraphContext'

interface State {}

export class Navbar extends React.Component<ParsedRouteProps, State> {
    public state: State = {}

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
                        // If on-prem, never show a user avatar or sign-in button
                        sourcegraphContext.onPrem ?
                            null :
                            sourcegraphContext.user ?
                                <Link to='/settings'><UserAvatar size={64} /></Link> :
                                <Link to='/sign-in' className='btn btn-primary'>Sign in</Link>
                    }
                </div>
            </div>
        )
    }
}
