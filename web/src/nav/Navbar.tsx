import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { enableSearch2 } from '../search'
import { SearchBox } from '../search/SearchBox'
import { SearchNavbarItem as SearchNavbarItem2 } from '../search2/SearchNavbarItem'
import { UserAvatar } from '../settings/user/UserAvatar'

interface Props {
    history: H.History
    location: H.Location
}

interface State {}

const SearchBox12 = enableSearch2 ? SearchNavbarItem2 : SearchBox

export class Navbar extends React.Component<Props, State> {
    public state: State = {}

    public render(): JSX.Element | null {
        return (
            <div className={`navbar ${enableSearch2 ? 'navbar-search2' : ''}`}>
                <div className="navbar__left">
                    {/* ?hp forces link to go to homepage regardless of current search query */}
                    <Link to="/search?hp" className="navbar__logo-link">
                        <img className="navbar__logo" src="/.assets/img/sourcegraph-mark.svg" />
                    </Link>
                </div>
                <div className="navbar__search-box-container">
                    <SearchBox12 history={this.props.history} location={this.props.location} />
                </div>
                <div className="navbar__right">
                    {!window.context.onPrem && (
                        <a href="https://about.sourcegraph.com" className="navbar__nav-link">
                            About
                        </a>
                    )}
                    {(() => {
                        if (window.context.user) {
                            return (
                                <Link to="/settings">
                                    <UserAvatar size={64} />
                                </Link>
                            )
                        } else if (window.context.onPrem) {
                            return null
                        } else {
                            // If on-prem, don't show a sign-in button
                            return (
                                <Link
                                    to={`/sign-in?returnTo=${encodeURIComponent(this.props.location.pathname)}`}
                                    className="btn btn-primary"
                                >
                                    Sign in
                                </Link>
                            )
                        }
                    })()}
                </div>
            </div>
        )
    }
}
