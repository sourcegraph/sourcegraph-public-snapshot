
import * as React from 'react'
import { Link } from 'react-router-dom'
import { SearchBox } from 'sourcegraph/search/SearchBox'
import { ParsedRouteProps } from 'sourcegraph/util/routes'

export class Navbar extends React.Component<ParsedRouteProps, {}> {
    public render(): JSX.Element | null {
        return (
            <div className='navbar'>
                <div className='navbar__left'>
                    <Link to='/' className='navbar__logo-link'>
                        <img className='navbar__logo' src='/.assets/img/sourcegraph-mark.svg' />
                    </Link>
                </div>
                <div className='navbar__search-box-container'>
                    <SearchBox {...this.props} />
                </div>
                <div className='navbar__right'></div>
            </div>
        )
    }
}
