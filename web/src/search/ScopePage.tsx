import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import { Repo as RepositoryIcon } from '@sourcegraph/icons/lib/Repo'
import marked from 'marked'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { combineLatest } from 'rxjs/observable/combineLatest'
import { of } from 'rxjs/observable/of'
import { catchError } from 'rxjs/operators/catchError'
import { concat } from 'rxjs/operators/concat'
import { map } from 'rxjs/operators/map'
import { switchMap } from 'rxjs/operators/switchMap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { RepoFileLink } from '../components/RepoFileLink'
import { SearchScope } from '../schema/settings.schema'
import { fetchReposByQuery } from '../search/backend'
import { submitSearch } from '../search/helpers'
import { QueryInput } from '../search/QueryInput'
import { queryUpdates } from '../search/QueryInput'
import { SearchButton } from '../search/SearchButton'
import { currentConfiguration } from '../settings/configuration'
import { eventLogger } from '../tracking/eventLogger'

const ScopeNotFound = () => (
    <HeroPage
        icon={DirectionalSignIcon}
        title="404: Not Found"
        subtitle="Sorry, we can&#39;t find a scope here. Add an ID and description to your search scope to get a landing page."
    />
)

interface ScopePageProps extends RouteComponentProps<{ id: GQLID }> {}

interface State {
    query: string
    repoList?: string[]
    searchScopes?: SearchScope[]
    id?: string
    name?: string
    value: string
    description?: string
    errorMessage?: string
}

export class ScopePage extends React.Component<ScopePageProps, State> {
    private subscriptions = new Subscription()
    private propUpdates = new Subject<ScopePageProps>()
    public state: State = {
        query: '',
        repoList: [],
        value: '',
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('Scope')

        this.subscriptions.add(
            combineLatest(
                this.propUpdates,
                currentConfiguration.pipe(map((config): SearchScope[] => config['search.scopes'] || []))
            )
                .pipe(
                    switchMap(([props, searchScopes]) => {
                        this.setState({ searchScopes })
                        const matchedScope = searchScopes.find(o => o.id === props.match.params.id)
                        if (matchedScope) {
                            queryUpdates.next(this.state.value)
                            if (matchedScope.value.includes('repo:') || matchedScope.value.includes('repogroup:')) {
                                return of(matchedScope).pipe(
                                    concat(
                                        fetchReposByQuery(matchedScope.value).pipe(
                                            map(repoList => ({ repoList, errorMessage: undefined })),
                                            catchError(err => {
                                                console.error(err)
                                                return [{ errorMessage: err.message, repoList: [] }]
                                            })
                                        )
                                    )
                                )
                            }
                            queryUpdates.next(this.state.value)
                            return [{ ...matchedScope, repoList: [], errorMessage: undefined }]
                        }
                        return [
                            {
                                id: undefined,
                                name: undefined,
                                value: '',
                                description: undefined,
                                repoList: [],
                                errorMessage: undefined,
                            },
                        ]
                    })
                )
                .subscribe(state => this.setState(state as State))
        )
        this.propUpdates.next(this.props)
    }

    public componentWillReceiveProps(newProps: ScopePageProps): void {
        this.propUpdates.next(newProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.searchScopes) {
            return null
        }

        if (!this.state.searchScopes.some((element: SearchScope) => element.id === this.props.match.params.id)) {
            return <ScopeNotFound />
        }
        const sanitizedMarkdown = marked(this.state.description || '', { gfm: true, breaks: true, sanitize: true })
        return (
            <div className="scope-page">
                <div className="scope-page__container">
                    <header>
                        <h1 className="scope-page__title">{this.state.name}</h1>
                        {this.state.description && (
                            <div
                                className="scope-page__section"
                                dangerouslySetInnerHTML={{ __html: sanitizedMarkdown }}
                            />
                        )}
                    </header>
                    <section>
                        <form className="scope-page__section-search" onSubmit={this.onSubmit}>
                            <div className="scope-page__input-scope">
                                <span className="scope-page__input-scope-text">{this.state.value}</span>
                            </div>
                            <QueryInput
                                value={this.state.query}
                                onChange={this.onQueryChange}
                                prependQueryForSuggestions={this.state.value}
                                autoFocus={true}
                                location={this.props.location}
                                history={this.props.history}
                                placeholder="Search in this scope..."
                            />
                            <SearchButton />
                        </form>
                    </section>
                    <PageTitle title={this.state.name} />
                    <section className="scope-page__repos">
                        <div>
                            <div>
                                {this.state.errorMessage && (
                                    <p className="alert alert-danger">{this.state.errorMessage}</p>
                                )}
                                {!this.state.errorMessage &&
                                    (this.state.repoList && this.state.repoList.length > 0 ? (
                                        <div>
                                            <p>Repositories included in this scope:</p>
                                            <div>
                                                {this.state.repoList.slice(0, 50).map((repo, i) => (
                                                    <div key={i} className="scope-page__row">
                                                        <Link to={`/${repo}`} className="scope-page__link">
                                                            <RepositoryIcon className="icon-inline scope-page__link-icon" />
                                                            <RepoFileLink repoPath={repo} disableLinks={true} />
                                                        </Link>
                                                    </div>
                                                ))}
                                            </div>
                                            <p className="scope-page__count">
                                                {this.state.repoList.length}{' '}
                                                {this.state.repoList.length > 1 ? 'repositories ' : 'repository '} total{' '}
                                                {this.state.repoList.length > 50 ? '(showing first 50)' : ''}{' '}
                                            </p>
                                        </div>
                                    ) : (
                                        <p>All repositories included in this scope</p>
                                    ))}
                            </div>
                            {window.context.sourcegraphDotComMode &&
                                !window.context.user && (
                                    <small>
                                        Have an idea for a search scope for your community, or want to add a repository
                                        to this scope? Tweet us <a href="https://twitter.com/srcgraph">@srcgraph</a>.
                                    </small>
                                )}
                        </div>
                    </section>
                </div>
            </div>
        )
    }

    private onQueryChange = (query: string) => this.setState({ query })

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        submitSearch(this.props.history, { query: `${this.state.value} ${this.state.query}` }, 'home')
    }
}
