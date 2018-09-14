import * as H from 'history'
import { escapeRegExp } from 'lodash'
import * as path from 'path'
import * as React from 'react'
import { matchPath } from 'react-router'
import { NavLink } from 'react-router-dom'
import { Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, filter, map, mergeMap } from 'rxjs/operators'
import { authRequired } from '../../auth'
import { currentUser } from '../../auth'
import * as GQL from '../../backend/graphqlschema'
import { Tooltip } from '../../components/tooltip/Tooltip'
import { routes } from '../../routes'
import { currentConfiguration } from '../../settings/configuration'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchSearchScopes } from '../backend'
import { FilterChip } from '../FilterChip'
import { submitSearch, toggleSearchFilter } from '../helpers'

interface Props {
    location: H.Location
    history: H.History

    /**
     * The current query.
     */
    query: string
}

interface ISearchScope {
    name?: string
    value: string
}

interface State {
    /** All search scopes from configuration */
    configuredScopes?: ISearchScope[]
    /** All fetched search scopes */
    remoteScopes?: ISearchScope[]
    user: GQL.IUser | null
}

export class SearchFilterChips extends React.PureComponent<Props, State> {
    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        this.state = { user: null }

        this.subscriptions.add(
            authRequired
                .pipe(
                    filter(authRequired => !authRequired),
                    mergeMap(() => fetchSearchScopes()),
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
            <div className="search-filter-chips">
                {/* Filtering out empty strings because old configurations have "All repositories" with empty value, which is useless with new chips design. */}
                {scopes
                    .filter(scope => scope.value !== '')
                    .map((scope, i) => (
                        <FilterChip
                            query={this.props.query}
                            onFilterChosen={this.onSearchScopeClicked}
                            key={i}
                            value={scope.value}
                            name={scope.name}
                        />
                    ))}
                {this.state.user && (
                    <div className="search-filter-chips__edit">
                        <NavLink
                            className="search-filter-chips__add-edit"
                            to="/settings"
                            data-tooltip={scopes.length > 0 ? 'Edit search scopes' : undefined}
                        >
                            <small className="search-filter-chips__center">
                                {scopes.length === 0 ? (
                                    <span className="search-filter-chips__add-scopes">
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
            const match = matchPath<{ repoRev?: string; filePath?: string }>(this.props.location.pathname, {
                path: route.path,
                exact: route.exact,
            })
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
                            value: `repo:^${escapeRegExp(repoPath)}$`,
                        })

                        if (match.params.filePath) {
                            const dirname = isTree ? match.params.filePath : path.dirname(match.params.filePath)
                            if (dirname !== '.') {
                                scopes.push({
                                    value: `repo:^${escapeRegExp(repoPath)}$ file:^${escapeRegExp(dirname)}/`,
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

    private onSearchScopeClicked = (value: string) => {
        eventLogger.log('SearchScopeClicked', {
            search_filter: {
                value,
            },
        })
        submitSearch(this.props.history, { query: toggleSearchFilter(this.props.query, value) }, 'filter')
    }
}

function scopeForRepo(repoPath: string): ISearchScope {
    return {
        value: `repo:^${escapeRegExp(repoPath)}$`,
    }
}
