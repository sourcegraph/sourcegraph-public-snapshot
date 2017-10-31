import * as H from 'history'
import * as React from 'react'
import { matchPath } from 'react-router'
import 'rxjs/add/operator/distinctUntilChanged'
import 'rxjs/add/operator/skip'
import 'rxjs/add/operator/startWith'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { routes } from '../routes'
import { viewEvents } from '../tracking/events'
import { submitSearch } from './helpers'
import { buildSearchURLQuery, parseSearchURLQuery } from './index'
import { QueryInput } from './QueryInput'
import { ScopeLabel } from './ScopeLabel'
import { SearchButton } from './SearchButton'
import { SearchScope } from './SearchScope'

interface Props {
    location: H.Location
    history: H.History
}

interface State {
    /** The query in the input field */
    userQuery: string

    /** The query value of the active search scope, or undefined if it's still loading */
    scopeQuery: string | undefined
}

/**
 * The search item in the navbar
 */
export class SearchNavbarItem extends React.Component<Props, State> {
    /** Subscriptions to unsubscribe from on component unmount */
    private subscriptions = new Subscription()

    /** Emits on componentWillReceiveProps */
    private componentUpdates = new Subject<Props>()

    constructor(props: Props) {
        super(props)

        // Fill text input from URL info
        this.state = this.getStateFromProps(props)

        /** Emits whenever the route changes */
        const routeChanges = this.componentUpdates
            .startWith(props)
            .distinctUntilChanged((a, b) => a.location === b.location)
            .skip(1)

        // Reset on route changes
        this.subscriptions.add(
            routeChanges.subscribe(props => {
                this.setState(this.getStateFromProps(props))
            }, err => {
                console.error(err)
            })
        )

        // Preserve search options ('q' and 'sq' query parameters) as the
        // user navigates, without requiring every <a href> value to contain
        // the URL query.
        this.subscriptions.add(
            routeChanges.subscribe(props => {
                const search = new URLSearchParams(props.location.search)

                let keepSearchOptionsParams = false
                for (const route of routes) {
                    const match = matchPath<{ repoRev?: string, filePath?: string }>(props.location.pathname, route)
                    if (match) {
                        switch (match.path) {
                            case '/:repoRev+': {
                                keepSearchOptionsParams = true
                                break
                            }
                            case '/:repoRev+/-/blob/:filePath+': {
                                keepSearchOptionsParams = true
                                break
                            }
                            case '/:repoRev+/-/tree/:filePath+': {
                                keepSearchOptionsParams = true
                                break
                            }
                            case '/search': {
                                keepSearchOptionsParams = false
                            }
                        }
                        break
                    }
                }
                if (!keepSearchOptionsParams) { return }

                if (!search.has('q') && !search.has('sq')) {
                    const searchOptions = parseSearchURLQuery(this.props.location.search)
                    if (searchOptions.query || searchOptions.scopeQuery) {
                        const urlWithSearchQueryParams = '?' + buildSearchURLQuery(searchOptions) + props.location.hash
                        props.history.replace(urlWithSearchQueryParams, props.location.state)
                    }
                }
            })
        )
    }

    public componentDidMount(): void {
        viewEvents.Home.log()
    }

    public componentWillReceiveProps(newProps: Props): void {
        this.componentUpdates.next(newProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        // Only autofocus the query input on search result pages (otherwise we
        // capture down-arrow keypresses that the user probably intends to scroll down
        // in the page).
        const autoFocus = this.props.location.pathname === '/search'

        return (
            <form
                className='search2 search-navbar-item2'
                onSubmit={this.onSubmit}
            >
                <div className='search-navbar-item2__row'>
                    <QueryInput
                        {...this.props}
                        value={this.state.userQuery}
                        onChange={this.onUserQueryChange}
                        scopeQuery={this.state.scopeQuery}
                        autoFocus={autoFocus ? 'cursor-at-end' : undefined}
                    />
                    <SearchButton />
                </div>
                <div className='search-navbar-item2__row'>
                    <SearchScope location={this.props.location} value={this.state.scopeQuery} onChange={this.onScopeQueryChange} />
                    <ScopeLabel scopeQuery={this.state.scopeQuery} />
                </div>
            </form>
        )
    }

    private onUserQueryChange = (userQuery: string) => {
        this.setState({ userQuery })
    }

    private onScopeQueryChange = (scopeQuery: string) => {
        this.setState({ scopeQuery })
    }

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        submitSearch(this.props.history, {
            query: this.state.userQuery,
            scopeQuery: this.state.scopeQuery || '',
        })
    }

    /**
     * Reads initial state from the props (i.e. URL parameters).
     */
    private getStateFromProps(props: Props): State {
        const options = parseSearchURLQuery(props.location.search || '')
        return {
            userQuery: options.query,
            scopeQuery: options.scopeQuery,
        }
    }
}
