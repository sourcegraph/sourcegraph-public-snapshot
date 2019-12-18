import * as H from 'history'
import { escapeRegExp } from 'lodash'
import * as path from 'path'
import * as React from 'react'
import { matchPath } from 'react-router'
import { NavLink } from 'react-router-dom'
import { Subject, Subscription } from 'rxjs'
import { distinctUntilChanged } from 'rxjs/operators'
import * as GQL from '../../../../shared/src/graphql/schema'
import { isSettingsValid, SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { Tooltip } from '../../components/tooltip/Tooltip'
import { routes } from '../../routes'
import { Settings } from '../../schema/settings.schema'
import { eventLogger } from '../../tracking/eventLogger'
import { FilterChip } from '../FilterChip'
import { submitSearch, toggleSearchFilter, toggleSearchFilterAndReplaceSampleRepogroup } from '../helpers'
import { PatternTypeProps } from '..'

interface Props extends SettingsCascadeProps, Omit<PatternTypeProps, 'setPatternType'> {
    location: H.Location
    history: H.History
    authenticatedUser: GQL.IUser | null
    isSourcegraphDotCom: boolean

    /**
     * The current query.
     */
    query: string
}

export interface ISearchScope {
    name?: string
    value: string
}

export class SearchFilterChips extends React.PureComponent<Props> {
    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Update tooltip text immediately after clicking.
        this.subscriptions.add(
            this.componentUpdates
                .pipe(distinctUntilChanged((a, b) => a.query === b.query))
                .subscribe(() => Tooltip.forceUpdate())
        )
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
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
                {this.props.authenticatedUser && (
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
                                    'Edit'
                                )}
                            </small>
                        </NavLink>
                    </div>
                )}
            </div>
        )
    }

    private getScopes(): ISearchScope[] {
        const allScopes: ISearchScope[] =
            (isSettingsValid<Settings>(this.props.settingsCascade) &&
                this.props.settingsCascade.final['search.scopes']) ||
            []
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
                        const [repoName] = match.params.repoRev!.split('@')
                        scopes.push(scopeForRepo(repoName))
                        break
                    }
                    case '/:repoRev+/-/tree/:filePath+':
                    case '/:repoRev+/-/blob/:filePath+': {
                        // Blob/tree page
                        const isTree = match.path === '/:repoRev+/-/tree/:filePath+'

                        const [repoName] = match.params.repoRev!.split('@')

                        scopes.push({
                            value: `repo:^${escapeRegExp(repoName)}$`,
                        })

                        if (match.params.filePath) {
                            const dirname = isTree ? match.params.filePath : path.dirname(match.params.filePath)
                            if (dirname !== '.') {
                                scopes.push({
                                    value: `repo:^${escapeRegExp(repoName)}$ file:^${escapeRegExp(dirname)}/`,
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

    private onSearchScopeClicked = (value: string): void => {
        eventLogger.log('SearchScopeClicked', {
            search_filter: {
                value,
            },
        })

        const newQuery = this.props.isSourcegraphDotCom
            ? toggleSearchFilterAndReplaceSampleRepogroup(this.props.query, value)
            : toggleSearchFilter(this.props.query, value)

        submitSearch(this.props.history, newQuery, 'filter', this.props.patternType)
    }
}

function scopeForRepo(repoName: string): ISearchScope {
    return {
        value: `repo:^${escapeRegExp(repoName)}$`,
    }
}
