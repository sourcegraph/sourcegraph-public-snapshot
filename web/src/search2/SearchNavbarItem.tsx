import * as H from 'history'
import * as React from 'react'
import { matchPath } from 'react-router'
import 'rxjs/add/operator/debounceTime'
import 'rxjs/add/operator/distinctUntilChanged'
import 'rxjs/add/operator/skip'
import 'rxjs/add/operator/startWith'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { routes } from '../routes'
import { Help } from './Help'
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

    /**
     * Emits when the scope query changes.
     */
    private scopeQueryChanges = new Subject<string>()

    constructor(props: Props) {
        super(props)

        // Fill text input from URL info
        this.state = this.getStateFromProps(props) || { userQuery: '' }

        this.subscriptions.add(
            this.scopeQueryChanges
                .distinctUntilChanged()
                .debounceTime(100)
                .subscribe(() => this.submit())
        )

        /** Emits whenever the route changes */
        const routeChanges = this.componentUpdates
            .startWith(props)
            .distinctUntilChanged((a, b) => a.location === b.location)
            .skip(1)

        // Reset on route changes
        this.subscriptions.add(
            routeChanges.subscribe(
                props => {
                    this.setState(this.getStateFromProps(props))
                },
                err => {
                    console.error(err)
                }
            )
        )

        // Listen to location changes in both ways. Depending on the source of the
        // history event, it might be seen first by one or the other. If we don't
        // listen for both, then we might receive some events too late.
        this.subscriptions.add(routeChanges.subscribe(props => this.onLocationChange(props.location)))
        this.subscriptions.add(props.history.listen(location => this.onLocationChange(location)))
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
            <form className="search2 search-navbar-item2" onSubmit={this.onSubmit}>
                <div className="search-navbar-item2__row">
                    <QueryInput
                        {...this.props}
                        value={this.state.userQuery}
                        onChange={this.onUserQueryChange}
                        scopeQuery={this.state.scopeQuery}
                        autoFocus={autoFocus ? 'cursor-at-end' : undefined}
                    />
                    <SearchButton />
                    <Help />
                </div>
                <div className="search-navbar-item2__row">
                    <SearchScope
                        location={this.props.location}
                        value={this.state.scopeQuery}
                        onChange={this.onScopeQueryChange}
                    />
                    <ScopeLabel scopeQuery={this.state.scopeQuery} />
                </div>
            </form>
        )
    }

    private onLocationChange = (location: H.Location): void => {
        // Preserve search options ('q' and 'sq' query parameters) as the
        // user navigates, without requiring every <a href> value to contain
        // the URL query.
        const search = new URLSearchParams(location.search)

        if (!this.keepSearchOptionsParams(location)) {
            return
        }

        if (!search.has('q') && !search.has('sq')) {
            const searchOptions = parseSearchURLQuery(this.props.location.search)
            if (searchOptions.query || searchOptions.scopeQuery) {
                const urlWithSearchQueryParams = '?' + buildSearchURLQuery(searchOptions) + location.hash
                this.props.history.replace(urlWithSearchQueryParams, location.state)
            }
        }
    }

    private keepSearchOptionsParams(location: H.Location): boolean {
        for (const route of routes) {
            const match = matchPath<{ repoRev?: string; filePath?: string }>(location.pathname, route)
            if (match) {
                switch (match.path) {
                    case '/:repoRev+':
                        return true

                    case '/:repoRev+/-/blob/:filePath+':
                        return true

                    case '/:repoRev+/-/tree/:filePath+':
                        return true

                    case '/search': {
                        // Interpret /search?hp as "go to homepage and clear query". Without this,
                        // we have no way of determining whether /search should go to the homepage
                        // with a blank query or to the search results page with the current query,
                        // since we attempt to preserve the current query across navigation in
                        // general.
                        const forceGoToHomepage = new URLSearchParams(location.search).has('hp')
                        return !forceGoToHomepage
                    }
                }
            }
        }

        return true
    }

    private onUserQueryChange = (userQuery: string) => {
        this.setState({ userQuery })
    }

    private onScopeQueryChange = (scopeQuery: string, needsDebounce?: boolean) => {
        this.setState({ scopeQuery }, () => this.scopeQueryChanges.next(scopeQuery))
    }

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        this.submit()
    }

    private submit(): void {
        submitSearch(this.props.history, {
            query: this.state.userQuery,
            scopeQuery: this.state.scopeQuery || '',
        })
    }

    /**
     * Reads initial state from the props (i.e. URL parameters).
     */
    private getStateFromProps(props: Props): State {
        if (this.keepSearchOptionsParams(props.location)) {
            const options = parseSearchURLQuery(props.location.search || '')
            const noQuery = !options.query && !options.scopeQuery
            return {
                userQuery: options.query,
                scopeQuery: noQuery ? undefined : options.scopeQuery,
            }
        }
        return this.state
    }
}
