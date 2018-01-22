import escapeRegexp from 'escape-string-regexp'
import * as H from 'history'
import * as path from 'path'
import * as React from 'react'
import { matchPath } from 'react-router'
import { NavLink } from 'react-router-dom'
import { catchError } from 'rxjs/operators/catchError'
import { map } from 'rxjs/operators/map'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from '../auth'
import { routes } from '../routes'
import { currentConfiguration } from '../settings/configuration'
import { eventLogger } from '../tracking/eventLogger'
import { fetchSearchScopes } from './backend'

interface Props {
    location: H.Location

    /**
     * The current query.
     */
    query: string

    /**
     * Called when there is a suggestion to be added to the search query.
     */
    onSuggestionChosen: (query: string) => void
}

interface ISearchScope {
    name: string
    value: string
}

/** The subset of State that is persisted to localStorage */
interface PersistableState {
    /** All fetched search scopes */
    remoteScopes?: ISearchScope[]
}

/** Data that is persisted to localStorage but NOT in the component state. */
interface PersistedState extends PersistableState {
    /** The value of the last-active scope. */
    lastScopeValue?: string
}

interface State extends PersistableState {
    /** All search scopes from configuration */
    configuredScopes?: ISearchScope[]
    user: GQL.IUser | null
}

export class SearchSuggestionChips extends React.PureComponent<Props, State> {
    private static REMOTE_SCOPES_STORAGE_KEY = 'SearchScope/remoteScopes'
    private static LAST_SCOPE_STORAGE_KEY = 'SearchScope/lastScope'

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        const savedState = this.loadFromLocalStorage()
        this.state = { remoteScopes: savedState.remoteScopes, user: null }

        // Always start with the scope suggestion that the user last clicked, if any.
        if (savedState.lastScopeValue) {
            this.props.onSuggestionChosen(savedState.lastScopeValue)
        } else if (window.context.sourcegraphDotComMode) {
            this.props.onSuggestionChosen('repogroup:sample ')
        }

        this.subscriptions.add(
            fetchSearchScopes()
                .pipe(
                    catchError(err => {
                        console.error(err)
                        return []
                    }),
                    map((remoteScopes: GQL.ISearchScope[]) => ({ remoteScopes }))
                )
                .subscribe(
                    newState =>
                        this.setState(newState, () => {
                            this.saveToLocalStorage()
                        }),
                    err => console.error(err)
                )
        )
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            currentConfiguration.pipe(map(config => config['search.scopes'] || [])).subscribe(searchScopes =>
                this.setState({
                    configuredScopes: searchScopes,
                })
            )
        )
        this.subscriptions.add(currentUser.subscribe(user => this.setState({ user })))
    }

    public componentWillReceiveProps(newProps: Props): void {
        this.componentUpdates.next(newProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const scopes = this.getScopes()

        return (
            <div className="search-suggestion-chips">
                {/* Filtering out empty strings because old configurations have "All repositories" with empty value, which is useless with new chips design. */}
                {scopes.filter(scope => scope.value !== '').map((scope, i) => (
                    <button
                        className="btn btn-secondary btn-sm search-suggestion-chips__chip"
                        key={i}
                        value={scope.value}
                        data-tooltip={this.props.query.includes(scope.value) ? 'Scope already in query' : scope.value}
                        disabled={this.props.query.includes(scope.value)}
                        onMouseDown={this.onMouseDown}
                        onClick={this.onClick}
                    >
                        {scope.name}
                    </button>
                ))}
                {this.state.user && (
                    <div className="search-suggestion-chips__edit">
                        <NavLink
                            className="search-page__edit"
                            to="/settings/configuration"
                            data-tooltip="Edit search scopes"
                        >
                            <small className="search-page__center">Edit</small>
                        </NavLink>
                    </div>
                )}
            </div>
        )
    }

    private onMouseDown: React.MouseEventHandler<HTMLButtonElement> = event => {
        // prevent clicking on chips from taking focus away from the search input.
        event.preventDefault()
    }

    private onClick: React.MouseEventHandler<HTMLButtonElement> = event => {
        eventLogger.log('SearchSuggestionClicked')
        event.preventDefault()
        this.props.onSuggestionChosen(event.currentTarget.value)

        // Persist the clicked suggestion to localstorage.
        this.saveToLocalStorage(this.props, event.currentTarget.value)
    }

    private getScopes(): ISearchScope[] {
        const allScopes: ISearchScope[] = []

        if (this.state.remoteScopes) {
            allScopes.push(...this.state.remoteScopes)
        }

        if (this.state.configuredScopes) {
            allScopes.push(...this.state.configuredScopes)
        }

        allScopes.push(...this.getScopesForCurrentRoute())
        return allScopes
    }

    /**
     * Returns contextual scopes for the current route (such as "This Repository" and
     * "This Directory").
     */
    private getScopesForCurrentRoute(): ISearchScope[] {
        const scopes: ISearchScope[] = []

        // This is basically a programmatical <Switch> with <Route>s
        // see https://reacttraining.com/react-router/web/api/matchPath
        // and https://reacttraining.com/react-router/web/example/sidebar
        for (const route of routes) {
            const match = matchPath<{ repoRev?: string; filePath?: string }>(this.props.location.pathname, route)
            if (match) {
                switch (match.path) {
                    case '/:repoRev+': {
                        // Repo page
                        const [repoPath] = match.params.repoRev!.split('@')
                        scopes.push(scopeForRepo(repoPath))
                        break
                    }
                    case '/:repoRev+/-/tree/:filePath+':
                    case '/:repoRev+/-/blob/:filePath+': {
                        // Blob/tree page
                        const isTree = match.path === '/:repoRev+/-/tree/:filePath+'

                        const [repoPath] = match.params.repoRev!.split('@')

                        scopes.push({
                            name: `This repository (${path.basename(repoPath)})`,
                            value: `repo:^${escapeRegexp(repoPath)}$`,
                        })

                        if (match.params.filePath) {
                            const dirname = isTree ? match.params.filePath : path.dirname(match.params.filePath)
                            if (dirname !== '.') {
                                scopes.push({
                                    name: `This directory (${path.basename(dirname)})`,
                                    value: `repo:^${escapeRegexp(repoPath)}$ file:^${escapeRegexp(dirname)}/`,
                                })
                            }
                        }
                        break
                    }
                }
                break
            }
        }

        return scopes
    }

    private saveToLocalStorage(props: Props = this.props, clickedSuggestion?: string): void {
        const writeItem = (key: string, data: any): void => {
            if (data !== undefined && data !== null) {
                localStorage.setItem(key, JSON.stringify(data))
            } else {
                localStorage.removeItem(key)
            }
        }

        writeItem(SearchSuggestionChips.REMOTE_SCOPES_STORAGE_KEY, this.state.remoteScopes)

        // Persist the clicked suggestion, if any.
        if (clickedSuggestion) {
            writeItem(SearchSuggestionChips.LAST_SCOPE_STORAGE_KEY, clickedSuggestion)
        }
    }

    private loadFromLocalStorage(): PersistedState {
        const readItem = <T extends {}>(key: string, validate: (data: T) => boolean): T | undefined => {
            const raw = localStorage.getItem(key)
            if (raw === null) {
                return undefined
            }

            try {
                const data = JSON.parse(raw)
                if (data !== undefined && data !== null && validate(data)) {
                    return data
                }
            } catch (err) {
                /* noop */
            }

            // Else invalid data.
            localStorage.removeItem(key)
            return undefined
        }

        const validate = (data: ISearchScope): boolean =>
            typeof data.name === 'string' && typeof data.value === 'string'
        const remoteScopes = readItem<ISearchScope[]>(SearchSuggestionChips.REMOTE_SCOPES_STORAGE_KEY, data =>
            data.every(validate)
        )
        const lastScopeValue = readItem<string>(
            SearchSuggestionChips.LAST_SCOPE_STORAGE_KEY,
            s => typeof s === 'string'
        )
        return { remoteScopes, lastScopeValue }
    }
}

function scopeForRepo(repoPath: string): ISearchScope {
    return {
        name: `This repository (${path.basename(repoPath)})`,
        value: `repo:^${escapeRegexp(repoPath)}$`,
    }
}
