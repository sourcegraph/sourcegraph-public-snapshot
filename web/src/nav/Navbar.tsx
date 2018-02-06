import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { queryIndexOfScope } from '../search/helpers'
import { parseSearchURLQuery, SearchOptions } from '../search/index'
import { SearchNavbarItem } from '../search/SearchNavbarItem'
import { SearchSuggestionChips } from '../search/SearchSuggestionChips'
import { eventLogger } from '../tracking/eventLogger'
import { NavLinks } from './NavLinks'

interface Props {
    history: H.History
    location: H.Location
    isLightTheme: boolean
    onThemeChange: () => void
}

interface State {
    userQuery: string
    showScopes: boolean
}

const SHOW_SCOPES_LOCAL_STORAGE_KEY = 'show-scopes'

export class Navbar extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)

        /**
         * Reads initial state from the props (i.e. URL parameters).
         */
        const options = parseSearchURLQuery(props.location.search || '')
        if (options) {
            this.state = {
                userQuery: options.query,
                showScopes: localStorage.getItem(SHOW_SCOPES_LOCAL_STORAGE_KEY) === 'true',
            }
        } else {
            // If we have no component state, then we may have gotten unmounted during a route change.
            const state: SearchOptions | undefined = props.location.state
            this.state = {
                userQuery: state ? state.query : '',
                showScopes: localStorage.getItem(SHOW_SCOPES_LOCAL_STORAGE_KEY) === 'true',
            }
        }
    }

    public componentDidUpdate(): void {
        localStorage.setItem(SHOW_SCOPES_LOCAL_STORAGE_KEY, this.state.showScopes + '')
    }

    public render(): JSX.Element | null {
        return (
            <div className="navbar">
                <div className="navbar__search">
                    <div className="navbar__left">
                        <Link to="/search" className="navbar__logo-link">
                            <img className="navbar__logo" src="/.assets/img/sourcegraph-mark.svg" />
                        </Link>
                    </div>
                    <div className="navbar__search-box-container">
                        <SearchNavbarItem
                            {...this.props}
                            userQuery={this.state.userQuery}
                            onChange={this.onUserQueryChange}
                        />
                    </div>
                    <NavLinks {...this.props} onShowScopes={this.onShowScopes} showScopes={this.state.showScopes} />
                </div>
                <div className={'navbar__scopesbar' + (this.state.showScopes ? '' : ' navbar__scopesbar--hidden')}>
                    <SearchSuggestionChips
                        location={this.props.location}
                        onSuggestionChosen={this.onSuggestionChosen}
                        query={this.state.userQuery}
                    />
                </div>
            </div>
        )
    }

    private onShowScopes = () => {
        eventLogger.log('ScopesBarToggled', { code_search: { scopes_bar_toggled: !this.state.showScopes } })
        this.setState(({ showScopes }) => ({ showScopes: !showScopes }))
    }

    private onUserQueryChange = (userQuery: string) => {
        this.setState({ userQuery })
    }

    private onSuggestionChosen = (scope: string): void => {
        const idx = queryIndexOfScope(this.state.userQuery, scope)
        if (idx === -1) {
            this.addScopeToQuery(scope)
        } else {
            this.removeScopeFromQuery(scope, idx)
        }
    }

    private addScopeToQuery(scope: string): void {
        this.setState(state => ({ userQuery: [state.userQuery.trim(), scope].filter(s => s).join(' ') + ' ' }))
    }

    private removeScopeFromQuery(scope: string, idx: number): void {
        this.setState(state => ({
            userQuery: (
                state.userQuery.substring(0, idx).trim() +
                ' ' +
                state.userQuery.substring(idx + scope.length).trim()
            ).trim(),
        }))
    }
}
