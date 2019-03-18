import marked from 'marked'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { combineLatest, of, Subject, Subscription } from 'rxjs'
import { catchError, concat, map, switchMap } from 'rxjs/operators'
import * as GQL from '../../backend/graphqlschema'
import { Form } from '../../components/Form'
import { HeroPage } from '../../components/HeroPage'
import { PageTitle } from '../../components/PageTitle'
import { RepoLink } from '../../repo/RepoLink'
import { SearchScope } from '../../schema/settings.schema'
import { currentConfiguration } from '../../settings/configuration'
import { eventLogger } from '../../tracking/eventLogger'
import { RepositoryIcon } from '../../util/icons' // TODO: Switch to mdi icon
import { fetchReposByQuery } from '../backend'
import { submitSearch } from '../helpers'
import { queryUpdates } from './QueryInput'
import { QueryInput } from './QueryInput'
import { SearchButton } from './SearchButton'

const ScopeNotFound = () => (
    <HeroPage
        icon={MapSearchIcon}
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

interface ScopePageProps extends RouteComponentProps<{ id: GQL.ID }> {
    authenticatedUser: GQL.IUser | null
}

interface State {
    query: string
    repositories?: { name: string; url: string }[]
    searchScopes?: SearchScope[]
    id?: string
    name?: string
    value: string
    markdownDescription?: string
    first: number
    errorMessage?: string
}

export class ScopePage extends React.Component<ScopePageProps, State> {
    private subscriptions = new Subscription()
    private propUpdates = new Subject<ScopePageProps>()
    private showMoreClicks = new Subject<void>()

    public state: State = {
        query: '',
        repositories: [],
        value: '',
        first: 50,
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
                                            map(repositories => ({ repositories, errorMessage: undefined })),
                                            catchError(err => {
                                                console.error(err)
                                                return [{ errorMessage: err.message, repositories: [], first: 0 }]
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
                                    repositories: [],
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
                                repositories: [],
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
                                    (this.state.repositories && this.state.repositories.length > 0 ? (
                                        <div>
                                            <p>Repositories included in this scope:</p>
                                            <div>
                                                {this.state.repositories.slice(0, this.state.first).map((repo, i) => (
                                                    <div key={i} className="scope-page__row">
                                                        <Link to={repo.url} className="scope-page__link">
                                                            <RepositoryIcon className="icon-inline scope-page__link-icon" />
                                                            <RepoLink repoPath={repo.name} to={null} />
                                                        </Link>
                                                    </div>
                                                ))}
                                            </div>
                                            <p className="scope-page__count">
                                                {this.state.repositories.length}{' '}
                                                {this.state.repositories.length > 1 ? 'repositories ' : 'repository '}{' '}
                                                total{' '}
                                                {this.state.repositories.length > this.state.first
                                                    ? `(showing first ${this.state.first})`
                                                    : ''}{' '}
                                            </p>
                                            {this.state.first < this.state.repositories.length && (
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
                                !this.props.authenticatedUser && (
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
