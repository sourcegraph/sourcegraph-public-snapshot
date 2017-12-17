import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { SearchNavbarItem } from '../search/SearchNavbarItem'
import { NavLinks } from './NavLinks'

interface Props {
    history: H.History
    location: H.Location
}

interface State {}

export class Navbar extends React.Component<Props, State> {
    public state: State = {}

    public render(): JSX.Element | null {
        return (
            <div className="navbar navbar-search">
                <div className="navbar__left">
                    <Link to="/search" className="navbar__logo-link">
                        <img className="navbar__logo" src="/.assets/img/sourcegraph-mark.svg" />
                    </Link>
                </div>
                <div className="navbar__search-box-container">
                    <SearchNavbarItem history={this.props.history} location={this.props.location} />
                </div>
                <NavLinks location={this.props.location} />
            </div>
        )
    }
}
