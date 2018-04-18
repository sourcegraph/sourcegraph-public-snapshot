import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../backend/graphqlschema'
import { parseSearchURLQuery, SearchOptions } from '../search/index'
import { SearchFilterChips } from '../search/input/SearchFilterChips'
import { SearchNavbarItem } from '../search/input/SearchNavbarItem'
import { eventLogger } from '../tracking/eventLogger'
import { NavLinks } from './NavLinks'

interface Props {
    history: H.History
    location: H.Location
    user: GQL.IUser | null
    isLightTheme: boolean
    onThemeChange: () => void
    navbarSearchQuery: string
    onNavbarQueryChange: (query: string) => void
    onFilterChosen: (value: string) => void
}

interface State {
    showScopes: boolean
}

const SHOW_SCOPES_LOCAL_STORAGE_KEY = 'show-scopes'

export class GlobalNavbar extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)

        /**
         * Reads initial state from the props (i.e. URL parameters).
         */
        const options = parseSearchURLQuery(props.location.search || '')
        if (options) {
            this.state = {
                showScopes: localStorage.getItem(SHOW_SCOPES_LOCAL_STORAGE_KEY) === 'true',
            }
            props.onNavbarQueryChange(options.query)
        } else {
            // If we have no component state, then we may have gotten unmounted during a route change.
            const state: SearchOptions | undefined = props.location.state
            this.state = {
                showScopes: localStorage.getItem(SHOW_SCOPES_LOCAL_STORAGE_KEY) === 'true',
            }
            props.onNavbarQueryChange(state ? state.query : '')
        }
    }

    public componentDidUpdate(): void {
        localStorage.setItem(SHOW_SCOPES_LOCAL_STORAGE_KEY, this.state.showScopes + '')
    }

    public render(): JSX.Element | null {
        return (
            <div className="global-navbar">
                <div className="global-navbar__search">
                    <div className="global-navbar__left">
                        <Link to="/search" className="global-navbar__logo-link">
                            <img className="global-navbar__logo" src="/.assets/img/sourcegraph-mark.svg" />
                        </Link>
                    </div>
                    <div className="global-navbar__search-box-container">
                        <SearchNavbarItem
                            {...this.props}
                            navbarSearchQuery={this.props.navbarSearchQuery}
                            onChange={this.props.onNavbarQueryChange}
                        />
                    </div>
                    <NavLinks
                        {...this.props}
                        className="global-navbar__nav-links"
                        onShowScopes={this.onShowScopes}
                        showScopes={this.state.showScopes}
                    />
                </div>
                <div
                    className={
                        'global-navbar__scopesbar' + (this.state.showScopes ? '' : ' global-navbar__scopesbar--hidden')
                    }
                >
                    <SearchFilterChips
                        location={this.props.location}
                        onFilterChosen={this.props.onFilterChosen}
                        query={this.props.navbarSearchQuery}
                    />
                </div>
            </div>
        )
    }

    private onShowScopes = () => {
        eventLogger.log('ScopesBarToggled', { code_search: { scopes_bar_toggled: !this.state.showScopes } })
        this.setState(({ showScopes }) => ({ showScopes: !showScopes }))
    }
}
