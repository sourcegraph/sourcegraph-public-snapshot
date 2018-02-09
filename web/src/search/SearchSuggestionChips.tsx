import escapeRegexp from 'escape-string-regexp'
import * as H from 'history'
import * as path from 'path'
import * as React from 'react'
import { matchPath } from 'react-router'
import { NavLink } from 'react-router-dom'
import { catchError } from 'rxjs/operators/catchError'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from '../auth'
import { Tooltip } from '../components/tooltip/Tooltip'
import { routes } from '../routes'
import { currentConfiguration } from '../settings/configuration'
import { eventLogger } from '../tracking/eventLogger'
import { fetchSearchScopes } from './backend'
import { queryIndexOfScope } from './helpers'

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

interface State {
    /** All search scopes from configuration */
    configuredScopes?: ISearchScope[]
    /** All fetched search scopes */
    remoteScopes?: ISearchScope[]
    user: GQL.IUser | null
}

export class SearchSuggestionChips extends React.PureComponent<Props, State> {
    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        this.state = { user: null }

        this.subscriptions.add(
            fetchSearchScopes()
                .pipe(
                    catchError(err => {
                        console.error(err)
                        return []
                    }),
                    map((remoteScopes: GQL.ISearchScope[]) => ({ remoteScopes }))
                )
                .subscribe(newState => this.setState(newState), err => console.error(err))
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

        // Update tooltip text immediately after clicking.
        this.subscriptions.add(
            this.componentUpdates
                .pipe(distinctUntilChanged((a, b) => a.query === b.query))
                .subscribe(() => Tooltip.forceUpdate())
        )
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
                        className={
                            'btn btn-sm search-suggestion-chips__chip' +
                            (this.isScopeSelected(this.props.query, scope.value)
                                ? ' search-suggestion-chips__chip--selected'
                                : '')
                        }
                        key={i}
                        value={scope.value}
                        data-tooltip={
                            this.isScopeSelected(this.props.query, scope.value) ? 'Scope already in query' : scope.value
                        }
                        onMouseDown={this.onMouseDown}
                        onClick={this.onClick}
                    >
                        {scope.name}
                    </button>
                ))}
                {this.state.user && (
                    <div className="search-suggestion-chips__edit">
                        <NavLink
                            className="search-suggestion-chips__add-edit"
                            to="/settings/configuration"
                            data-tooltip={scopes.length > 0 ? 'Edit search scopes' : undefined}
                        >
                            <small className="search-suggestion-chips__center">
                                {scopes.length === 0 ? (
                                    <span className="search-suggestion-chips__add-scopes">
                                        Add search scopes for quick filtering
                                    </span>
                                ) : (
                                    `Edit`
                                )}
                            </small>
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
        eventLogger.log('SearchSuggestionClicked', { code_search: { suggestion_chip: event.currentTarget.value } })
        event.preventDefault()
        this.props.onSuggestionChosen(event.currentTarget.value)
    }

    private getScopes(): ISearchScope[] {
        const allScopes: ISearchScope[] = []

        if (this.state.configuredScopes) {
            allScopes.push(...this.state.configuredScopes)
        }

        if (this.state.remoteScopes) {
            allScopes.push(...this.state.remoteScopes)
        }

        allScopes.push(...this.getScopesForCurrentRoute())
        return allScopes
    }

    private isScopeSelected(query: string, scope: string): boolean {
        return queryIndexOfScope(query, scope) !== -1
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
}

function scopeForRepo(repoPath: string): ISearchScope {
    return {
        name: `This repository (${path.basename(repoPath)})`,
        value: `repo:^${escapeRegexp(repoPath)}$`,
    }
}
