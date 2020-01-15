import MapSearchIcon from 'mdi-react/MapSearchIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { concat, of, Subject, Subscription } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'
import { RepoLink } from '../../../../shared/src/components/RepoLink'
import * as GQL from '../../../../shared/src/graphql/schema'
import { isSettingsValid, SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { renderMarkdown } from '../../../../shared/src/util/markdown'
import { Form } from '../../components/Form'
import { HeroPage } from '../../components/HeroPage'
import { PageTitle } from '../../components/PageTitle'
import { SearchScope, Settings } from '../../schema/settings.schema'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchReposByQuery } from '../backend'
import { submitSearch, QueryState } from '../helpers'
import { QueryInput, queryUpdates } from './QueryInput'
import { SearchButton } from './SearchButton'
import { PatternTypeProps } from '..'
import { ErrorAlert } from '../../components/alerts'

const ScopeNotFound: React.FunctionComponent = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle={
            <>
                No search page found with that scope ID. Add an ID and description to a search scope to create a search
                page. See{' '}
                <Link to="/help/user/search/scopes#creating-search-scope-pages">search scope documentation</Link>.
            </>
        }
    />
)

interface ScopePageProps extends RouteComponentProps<{ id: GQL.ID }>, SettingsCascadeProps, PatternTypeProps {
    authenticatedUser: GQL.IUser | null
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

        that.subscriptions.add(
            that.propUpdates
                .pipe(
                    switchMap(props => {
                        const searchScopes =
                            (isSettingsValid<Settings>(props.settingsCascade) &&
                                props.settingsCascade.final['search.scopes']) ||
                            []
                        that.setState({ searchScopes })
                        const matchedScope = searchScopes.find(o => o.id === props.match.params.id)
                        if (matchedScope) {
                            const markdownDescription = renderMarkdown(matchedScope.description || '')
                            queryUpdates.next(matchedScope.value)
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
                .subscribe(state => that.setState(state as State))
        )

        that.subscriptions.add(
            that.showMoreClicks.subscribe(() => that.setState(state => ({ first: state.first + 50 })))
        )

        that.propUpdates.next(that.props)
    }

    public componentDidUpdate(): void {
        that.propUpdates.next(that.props)
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!that.state.searchScopes) {
            return null
        }

        if (!that.state.searchScopes.some((element: SearchScope) => element.id === that.props.match.params.id)) {
            return <ScopeNotFound />
        }
        return (
            <div className="scope-page">
                <div className="scope-page__container">
                    <header>
                        <h1 className="scope-page__title">{that.state.name}</h1>
                        {that.state.markdownDescription && (
                            <div
                                className="scope-page__section"
                                dangerouslySetInnerHTML={{ __html: that.state.markdownDescription || '' }}
                            />
                        )}
                    </header>
                    <section>
                        <Form className="scope-page__section-search" onSubmit={that.onSubmit}>
                            <div className="scope-page__input-scope" title={that.state.value}>
                                <span className="scope-page__input-scope-text">{that.state.value}</span>
                            </div>
                            <QueryInput
                                {...that.props}
                                value={that.state.queryState}
                                onChange={that.onQueryChange}
                                prependQueryForSuggestions={that.state.value}
                                autoFocus={true}
                                location={that.props.location}
                                history={that.props.history}
                                placeholder="Search in this scope..."
                            />
                            <SearchButton />
                        </Form>
                    </section>
                    <PageTitle title={that.state.name} />
                    <section className="scope-page__repos">
                        <div>
                            <div>
                                {that.state.errorMessage && <ErrorAlert error={that.state.errorMessage} />}
                                {!that.state.errorMessage &&
                                    (that.state.repositories && that.state.repositories.length > 0 ? (
                                        <div>
                                            <p>Repositories included in that scope:</p>
                                            <div>
                                                {that.state.repositories.slice(0, that.state.first).map((repo, i) => (
                                                    <div key={repo.name + String(i)} className="scope-page__row">
                                                        <Link to={repo.url} className="scope-page__link">
                                                            <SourceRepositoryIcon className="icon-inline scope-page__link-icon" />
                                                            <RepoLink repoName={repo.name} to={null} />
                                                        </Link>
                                                    </div>
                                                ))}
                                            </div>
                                            <p className="scope-page__count">
                                                {that.state.repositories.length}{' '}
                                                {that.state.repositories.length > 1 ? 'repositories ' : 'repository '}{' '}
                                                total{' '}
                                                {that.state.repositories.length > that.state.first
                                                    ? `(showing first ${that.state.first})`
                                                    : ''}{' '}
                                            </p>
                                            {that.state.first < that.state.repositories.length && (
                                                <button
                                                    type="button"
                                                    className="btn btn-secondary btn-sm scope-page__show-more"
                                                    onClick={that.onShowMore}
                                                >
                                                    Show more
                                                </button>
                                            )}
                                        </div>
                                    ) : (
                                        <p>All repositories included in that scope</p>
                                    ))}
                            </div>
                            {window.context.sourcegraphDotComMode && !that.props.authenticatedUser && (
                                <small>
                                    Have an idea for a search scope for your community, or want to add a repository to
                                    that scope? Tweet us <a href="https://twitter.com/srcgraph">@srcgraph</a>.
                                </small>
                            )}
                        </div>
                    </section>
                </div>
            </div>
        )
    }

    private onQueryChange = (queryState: QueryState): void => that.setState({ queryState })

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        submitSearch(
            that.props.history,
            `${that.state.value} ${that.state.queryState.query}`,
            'home',
            that.props.patternType
        )
    }

    private onShowMore = (): void => {
        that.showMoreClicks.next()
    }
}
