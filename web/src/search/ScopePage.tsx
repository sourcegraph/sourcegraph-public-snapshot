import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import { Repo as RepositoryIcon } from '@sourcegraph/icons/lib/Repo'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { combineLatest } from 'rxjs/observable/combineLatest'
import { map } from 'rxjs/operators/map'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { RepoBreadcrumb } from '../components/Breadcrumb'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { SearchScope } from '../schema/settings.schema'
import { fetchRepoGroups } from '../search/backend'
import { submitSearch } from '../search/helpers'
import { queryUpdates } from '../search/QueryInput'
import { QueryInput } from '../search/QueryInput'
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
    name: string
    value: string
    description?: string
}

export class ScopePage extends React.Component<ScopePageProps, State> {
    private subscriptions = new Subscription()
    private propUpdates = new Subject<ScopePageProps>()
    public state: State = {
        query: '',
        repoList: [],
        name: '',
        value: '',
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('Scope')

        this.subscriptions.add(
            combineLatest(
                fetchRepoGroups(),
                this.propUpdates,
                currentConfiguration.pipe(map((config): SearchScope[] => config['search.scopes'] || []))
            ).subscribe(([groups, props, searchScopes]) => {
                this.setState({ searchScopes })
                const matchedScope = searchScopes.find(o => o.id === props.match.params.id)
                if (matchedScope) {
                    this.setState({
                        id: matchedScope.id,
                        name: matchedScope.name,
                        value: matchedScope.value,
                        description: matchedScope.description,
                    })
                    const [, repoGroup] = matchedScope.value.split('repogroup:')
                    if (matchedScope.value && repoGroup) {
                        const repogroup = repoGroup.split(/\s/, 1)[0]
                        const scope = groups.find(o => o.name === repogroup)
                        this.setState({ repoList: scope ? scope.repositories : [] })
                    } else {
                        this.setState({ repoList: [] })
                    }
                    queryUpdates.next(this.state.value)
                } else {
                    this.setState({
                        id: undefined,
                        name: '',
                        value: '',
                        description: undefined,
                        repoList: [],
                    })
                }
            })
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
                        {this.state.description && <p>{this.state.description}</p>}
                    </header>
                    <section className="">
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
                        {this.state.repoList &&
                            this.state.repoList.length > 0 && <p>Repositories included in this scope:</p>}
                        <div>
                            <div>
                                {this.state.repoList && this.state.repoList.length > 0 ? (
                                    this.state.repoList.slice(0, 50).map((repo, i) => (
                                        <div key={i} className="scope-page__row">
                                            <Link to={`/${repo}`} className="scope-page__link">
                                                <RepositoryIcon className="icon-inline scope-page__link-icon" />
                                                <RepoBreadcrumb repoPath={repo} disableLinks={true} />
                                            </Link>
                                        </div>
                                    ))
                                ) : (
                                    <p>All repositories included in this scope</p>
                                )}
                            </div>
                            {this.state.repoList &&
                                this.state.repoList.length > 0 && (
                                    <p className="scope-page__count">
                                        {this.state.repoList.length} repositories total{' '}
                                        {this.state.repoList.length > 50 ? '(showing first 50)' : ''}{' '}
                                    </p>
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
