import GearIcon from '@sourcegraph/icons/lib/Gear'
import Loader from '@sourcegraph/icons/lib/Loader'
import format from 'date-fns/format'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { debounceTime } from 'rxjs/operators/debounceTime'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { startWith } from 'rxjs/operators/startWith'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { pluralize } from '../util/strings'
import { fetchAllRepositories } from './backend'

interface Props extends RouteComponentProps<any> {}

export interface State {
    query: string
    repos?: GQL.IRepository[]
    totalCount?: number
}

/**
 * A page displaying the repositories on this site.
 */
export class SiteAdminRepositoriesPage extends React.Component<Props, State> {
    private queryInputChanges = new Subject<string>()
    private subscriptions = new Subscription()

    public constructor(props: Props) {
        super(props)

        this.state = {
            query: new URLSearchParams(this.props.location.search).get('q') || '',
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminRepos')

        this.subscriptions.add(
            this.queryInputChanges
                .pipe(
                    startWith(this.state.query),
                    distinctUntilChanged(),
                    tap(query => this.setState({ query })),
                    debounceTime(500),
                    tap(query => this.props.history.replace({ search: `q=${encodeURIComponent(query)}` })),
                    switchMap(fetchAllRepositories)
                )
                .subscribe(resp =>
                    this.setState({
                        repos: resp.nodes,
                        totalCount: resp.totalCount,
                    })
                )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-detail-list site-admin-repositories-page">
                <PageTitle title="Repositories" />
                <h2>
                    Repositories{' '}
                    {typeof this.state.totalCount === 'number' &&
                        this.state.totalCount > 0 &&
                        `(${this.state.totalCount})`}
                </h2>
                <div className="site-admin-page__actions">
                    <Link
                        to="/site-admin/configuration"
                        className="btn btn-primary btn-sm site-admin-page__actions-btn"
                    >
                        <GearIcon className="icon-inline" /> Configure repositories
                    </Link>
                </div>
                <form className="site-admin-page__filter-form">
                    <input
                        className="form-control"
                        type="search"
                        placeholder="Search repositories..."
                        name="query"
                        value={this.state.query}
                        onChange={this.onChange}
                    />
                </form>
                {!this.state.repos && <Loader className="icon-inline" />}
                <ul className="site-admin-detail-list__list">
                    {this.state.repos &&
                        this.state.repos.map(repo => (
                            <li
                                key={repo.id}
                                className="site-admin-detail-list__item site-admin-repositories-page__repo"
                            >
                                <div className="site-admin-detail-list__header site-admin-repositories-page__repo-header">
                                    <Link to={`/${repo.uri}`} className="site-admin-detail-list__name">
                                        {repo.uri}
                                    </Link>
                                </div>
                                <ul className="site-admin-detail-list__info site-admin-repositories-page__repo-info">
                                    <li>ID: {repo.id}</li>
                                    {repo.createdAt && <li>Created: {format(repo.createdAt, 'YYYY-MM-DD')}</li>}
                                </ul>
                            </li>
                        ))}
                </ul>
                {this.state.repos &&
                    typeof this.state.totalCount === 'number' &&
                    (this.state.totalCount > 0 ? (
                        <p>
                            <small>
                                {this.state.totalCount} {pluralize('repository', this.state.totalCount, 'repositories')}{' '}
                                total{' '}
                                {this.state.repos.length < this.state.totalCount &&
                                    `(showing ${this.state.query ? 'matching' : 'first'} ${this.state.repos.length})`}
                            </small>
                        </p>
                    ) : (
                        <p>No repositories.</p>
                    ))}
            </div>
        )
    }

    private onChange: React.ChangeEventHandler<HTMLInputElement> = e => {
        this.queryInputChanges.next(e.currentTarget.value)
    }
}
