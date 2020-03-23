import MapSearchIcon from 'mdi-react/MapSearchIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { concat, of, Subject, Subscription } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'
import { RepoLink } from '../../../shared/src/components/RepoLink'
import * as GQL from '../../../shared/src/graphql/schema'
import { isSettingsValid, SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { renderMarkdown } from '../../../shared/src/util/markdown'
import { Form } from '../components/Form'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { SearchScope, Settings } from '../schema/settings.schema'
import { eventLogger } from '../tracking/eventLogger'
import { fetchReposByQuery } from './backend'
import { submitSearch, QueryState } from './helpers'
import { QueryInput } from './input/QueryInput'
import { SearchButton } from './input/SearchButton'
import { PatternTypeProps, CaseSensitivityProps } from '.'
import { ErrorAlert } from '../components/alerts'

const ScopeNotFound: React.FunctionComponent = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle={
            <>
                No search page found with this scope ID. Add an ID and description to a search scope to create a search
                page. See{' '}
                <Link to="/help/user/search/scopes#creating-search-scope-pages">search scope documentation</Link>.
            </>
        }
    />
)

interface ScopePageProps
    extends RouteComponentProps<{ id: GQL.ID }>,
        SettingsCascadeProps,
        PatternTypeProps,
        CaseSensitivityProps {
    authenticatedUser: GQL.IUser | null
    onNavbarQueryChange: (queryState: QueryState) => void
}

interface State {
    queryState: QueryState
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
        queryState: { query: '', cursorPosition: 0 },
        repositories: [],
        value: '',
        first: 50,
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('Scope')

        this.subscriptions.add(
            this.propUpdates
                .pipe(
                    switchMap(props => {
                        const searchScopes =
                            (isSettingsValid<Settings>(props.settingsCascade) &&
                                props.settingsCascade.final['search.scopes']) ||
                            []
                        this.setState({ searchScopes })
                        const matchedScope = searchScopes.find(o => o.id === props.match.params.id)
                        if (matchedScope) {
                            const markdownDescription = renderMarkdown(matchedScope.description || '')
                            this.props.onNavbarQueryChange({
                                query: matchedScope.value,
                                cursorPosition: matchedScope.value.length,
                            })
                            if (matchedScope.value.includes('repo:') || matchedScope.value.includes('repogroup:')) {
                                return concat(
                                    of({ ...matchedScope, markdownDescription }),
                                    fetchReposByQuery(matchedScope.value).pipe(
                                        map(repositories => ({ repositories, errorMessage: undefined })),
                                        catchError(err => {
                                            console.error(err)
                                            return [{ errorMessage: err.message, repositories: [], first: 0 }]
                                        })
                                    )
                                )
                            }
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

    public componentDidUpdate(): void {
        this.propUpdates.next(this.props)
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
                                {...this.props}
                                value={this.state.queryState}
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
                                {this.state.errorMessage && <ErrorAlert error={this.state.errorMessage} />}
                                {!this.state.errorMessage &&
                                    (this.state.repositories && this.state.repositories.length > 0 ? (
                                        <div>
                                            <p>Repositories included in this scope:</p>
                                            <div>
                                                {this.state.repositories.slice(0, this.state.first).map((repo, i) => (
                                                    <div key={repo.name + String(i)} className="scope-page__row">
                                                        <Link to={repo.url} className="scope-page__link">
                                                            <SourceRepositoryIcon className="icon-inline scope-page__link-icon" />
                                                            <RepoLink repoName={repo.name} to={null} />
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
                                                    type="button"
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
                            {window.context.sourcegraphDotComMode && !this.props.authenticatedUser && (
                                <small>
                                    Have an idea for a search scope for your community, or want to add a repository to
                                    this scope? Tweet us <a href="https://twitter.com/srcgraph">@srcgraph</a>.
                                </small>
                            )}
                        </div>
                    </section>
                </div>
            </div>
        )
    }

    private onQueryChange = (queryState: QueryState): void => this.setState({ queryState })

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        submitSearch(
            this.props.history,
            `${this.state.value} ${this.state.queryState.query}`,
            'home',
            this.props.patternType,
            this.props.caseSensitive
        )
    }

    private onShowMore = (): void => {
        this.showMoreClicks.next()
    }
}
