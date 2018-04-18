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
import { currentUser } from '../../auth'
import * as GQL from '../../backend/graphqlschema'
import { Form } from '../../components/Form'
import { HeroPage } from '../../components/HeroPage'
import { PageTitle } from '../../components/PageTitle'
import { RepoFileLink } from '../../components/RepoFileLink'
import { SearchScope } from '../../schema/settings.schema'
import { currentConfiguration } from '../../settings/configuration'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchReposByQuery } from '../backend'
import { submitSearch } from '../helpers'
import { queryUpdates } from '../input/QueryInput'
import { QueryInput } from '../input/QueryInput'
import { SearchButton } from '../input/SearchButton'

const ScopeNotFound = () => (
    <HeroPage
        icon={DirectionalSignIcon}
        title="404: Not Found"
        subtitle={
            <>
                No search page found with this scope ID. Add an ID and description to a search scope to create a search
                page. See{' '}
                <a href="https://about.sourcegraph.com/docs/server/config/search-scopes">search scope documentation</a>.
            </>
        }
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
    markdownDescription?: string
    first: number
    errorMessage?: string
    user?: GQL.IUser | null
}

export class ScopePage extends React.Component<ScopePageProps, State> {
    private subscriptions = new Subscription()
    private propUpdates = new Subject<ScopePageProps>()
    private showMoreClicks = new Subject<void>()

    public state: State = {
        query: '',
        repoList: [],
        value: '',
        first: 50,
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('Scope')
        this.subscriptions.add(currentUser.subscribe(user => this.setState({ user })))

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
                            const markdownDescription = marked(matchedScope.description || '', {
                                gfm: true,
                                breaks: true,
                                sanitize: true,
                            })
                            queryUpdates.next(matchedScope.value)
                            if (matchedScope.value.includes('repo:') || matchedScope.value.includes('repogroup:')) {
                                return of({
                                    ...matchedScope,
                                    markdownDescription,
                                }).pipe(
                                    concat(
                                        fetchReposByQuery(matchedScope.value).pipe(
                                            map(repoList => ({ repoList, errorMessage: undefined })),
                                            catchError(err => {
                                                console.error(err)
                                                return [{ errorMessage: err.message, repoList: [], first: 0 }]
                                            })
                                        )
                                    )
                                )
                            }
                            queryUpdates.next(matchedScope.value)
                            return [
                                {
                                    ...matchedScope,
                                    markdownDescription,
                                    repoList: [],
                                    errorMessage: undefined,
                                    first: 0,
                                },
                            ]
                        }
                        return [
                            {
                                id: undefined,
                                name: undefined,
                                value: '',
                                markdownDescription: '',
                                repoList: [],
                                errorMessage: undefined,
                                first: 0,
                            },
                        ]
                    })
                )
                .subscribe(state => this.setState(state as State))
        )

        this.subscriptions.add(
            this.showMoreClicks.subscribe(() => this.setState(state => ({ first: state.first + 50 })))
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
        return (
            <div className="scope-page">
                <div className="scope-page__container">
                    <header>
                        <h1 className="scope-page__title">{this.state.name}</h1>
                        {this.state.markdownDescription && (
                            <div
                                className="scope-page__section"
                                dangerouslySetInnerHTML={{ __html: this.state.markdownDescription || '' }}
                            />
                        )}
                    </header>
                    <section>
                        <Form className="scope-page__section-search" onSubmit={this.onSubmit}>
                            <div className="scope-page__input-scope" title={this.state.value}>
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
                        </Form>
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
                                                {this.state.repoList.slice(0, this.state.first).map((repo, i) => (
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
                                                {this.state.repoList.length > this.state.first
                                                    ? `(showing first ${this.state.first})`
                                                    : ''}{' '}
                                            </p>
                                            {this.state.first < this.state.repoList.length && (
                                                <button
                                                    className="btn btn-secondary btn-sm scope-page__show-more"
                                                    onClick={this.onShowMore}
                                                >
                                                    Show more
                                                </button>
                                            )}
                                        </div>
                                    ) : (
                                        <p>All repositories included in this scope</p>
                                    ))}
                            </div>
                            {window.context.sourcegraphDotComMode &&
                                !this.state.user && (
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

    private onShowMore = (event: React.MouseEvent<HTMLButtonElement>): void => {
        this.showMoreClicks.next()
    }
}
