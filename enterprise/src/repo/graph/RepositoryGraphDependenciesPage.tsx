import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import DotsHorizontalIcon from 'mdi-react/DotsHorizontalIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, filter, map, switchMap } from 'rxjs/operators'
import { gql, queryGraphQL } from '../../../../src/backend/graphql'
import * as GQL from '../../../../src/backend/graphqlschema'
import { FilteredConnection } from '../../../../src/components/FilteredConnection'
import { PageTitle } from '../../../../src/components/PageTitle'
import { buildSearchURLQuery, searchQueryForRepoRev } from '../../../../src/search'
import { eventLogger } from '../../../../src/tracking/eventLogger'
import { createAggregateError, ErrorLike, isErrorLike } from '../../../../src/util/errors'
import { pluralize } from '../../../../src/util/strings'

interface DependencyNodePropsCommon {
    repo: GQL.IRepository
    rev: string | undefined
    commitID: string
}

interface DependencyNodeProps extends DependencyNodePropsCommon {
    node: GQL.IDependency
}

interface DependencyNodeState {
    referenceCountOrError?: GQL.IApproximateCount | null | ErrorLike
}

class DependencyNode extends React.PureComponent<DependencyNodeProps, DependencyNodeState> {
    public state: DependencyNodeState = {}

    private componentUpdates = new Subject<DependencyNodeProps>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    map(({ node }) => node),
                    distinctUntilChanged(),
                    filter(node => !!node.references),
                    switchMap(node =>
                        this.fetchReferenceCount(node).pipe(
                            catchError(error => {
                                console.error(error)
                                this.setState({ referenceCountOrError: error })
                                return []
                            })
                        )
                    )
                )
                .subscribe(
                    references => this.setState({ referenceCountOrError: references }),
                    error => console.error(error)
                )
        )

        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(nextProps: DependencyNodeProps): void {
        this.componentUpdates.next(nextProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const node = this.props.node
        return (
            <li className="repository-graph-page__node">
                <div className="repository-graph-page__node-title">
                    {[{ key: 'language', value: node.language }, ...node.data]
                        .filter(({ key, value }) => !!value && key !== 'absolute')
                        .map(({ key, value }) => (
                            <span key={key} className="repository-graph-page__node-datum">
                                <span className="repository-graph-page__node-datum-name">{key}:</span>
                                <span className="repository-graph-page__node-datum-value">{String(value)}</span>
                            </span>
                        ))}
                </div>
                <div className="repository-graph-page__node-actions">
                    {node.references ? (
                        <>
                            {this.state.referenceCountOrError === undefined && (
                                <LoadingSpinner className="icon-inline" />
                            )}
                            {isErrorLike(this.state.referenceCountOrError) ? (
                                <span title={this.state.referenceCountOrError.message}>
                                    <AlertCircleIcon className="icon-inline" />
                                </span>
                            ) : (
                                <Link
                                    to={this.urlToSearch(node.references.queryString)}
                                    className="btn btn-link btn-sm"
                                >
                                    {this.state.referenceCountOrError === undefined ||
                                    this.state.referenceCountOrError === null ? (
                                        'References in this repository'
                                    ) : (
                                        <>
                                            {this.state.referenceCountOrError.label}{' '}
                                            {pluralize('reference', this.state.referenceCountOrError.count)} in this
                                            repository
                                        </>
                                    )}
                                </Link>
                            )}
                        </>
                    ) : (
                        <DotsHorizontalIcon
                            className="icon-inline repository-graph-page__node-dotdotdot"
                            data-tooltip={`Reference search is not supported by the language server (${
                                this.props.node.language
                            }).`}
                        />
                    )}
                </div>
            </li>
        )
    }

    private urlToSearch(partialQuery: string | null): string {
        if (partialQuery === null) {
            return ''
        }
        const query = `${searchQueryForRepoRev(this.props.repo.name, this.props.rev)}type:ref ${partialQuery}`
        return `/search?${buildSearchURLQuery({ query })}`
    }

    private fetchReferenceCount = (node: GQL.IDependency): Observable<GQL.IApproximateCount | null> =>
        queryGraphQL(
            gql`
                query DependencyReferenceCount($dependency: ID!) {
                    node(id: $dependency) {
                        ... on Dependency {
                            references {
                                approximateCount {
                                    count
                                    exact
                                    label
                                }
                            }
                        }
                    }
                }
            `,
            { dependency: node.id }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.node) {
                    throw createAggregateError(errors)
                }
                const dependency = data.node as GQL.IDependency
                if (!dependency.references) {
                    throw createAggregateError(errors)
                }
                return dependency.references.approximateCount
            })
        )
}

interface Props extends RouteComponentProps<{}> {
    repo: GQL.IRepository
    rev: string | undefined
    commitID: string
}

interface State {}

class FilteredDependencyConnection extends FilteredConnection<GQL.IDependency, DependencyNodePropsCommon> {}

/**
 * The repository graph dependencies page.
 */
export class RepositoryGraphDependenciesPage extends React.PureComponent<Props, State> {
    public state: State = {}

    private repoChanges = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('RepoGraphDependencies')
    }

    public componentWillReceiveProps(nextProps: Props): void {
        if (nextProps.repo.id !== this.props.repo.id) {
            this.repoChanges.next()
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="repository-graph-page">
                <PageTitle title="Dependencies - Repository graph" />
                <h2>Dependencies</h2>
                <FilteredDependencyConnection
                    noun="dependency"
                    pluralNoun="dependencies"
                    queryConnection={this.fetchDependencies}
                    nodeComponent={DependencyNode}
                    nodeComponentProps={
                        {
                            repo: this.props.repo,
                            rev: this.props.rev,
                            commitID: this.props.commitID,
                        } as DependencyNodePropsCommon
                    }
                    defaultFirst={10}
                    updates={this.repoChanges}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private fetchDependencies = (args: { first?: number; query?: string }): Observable<GQL.IDependencyConnection> =>
        queryGraphQL(
            gql`
                query RepositoryDependencies($repo: ID!, $commitID: String!, $first: Int, $query: String) {
                    node(id: $repo) {
                        ... on Repository {
                            commit(rev: $commitID) {
                                dependencies(first: $first, query: $query) {
                                    nodes {
                                        id
                                        language
                                        data {
                                            key
                                            value
                                        }
                                        hints {
                                            key
                                            value
                                        }
                                        references {
                                            queryString
                                        }
                                    }
                                    totalCount
                                    pageInfo {
                                        hasNextPage
                                    }
                                }
                            }
                        }
                    }
                }
            `,
            { ...args, repo: this.props.repo.id, commitID: this.props.commitID }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.node) {
                    throw createAggregateError(errors)
                }
                const repo = data.node as GQL.IRepository
                if (!repo.commit || !repo.commit.dependencies || !repo.commit.dependencies.nodes) {
                    throw createAggregateError(errors)
                }
                return repo.commit.dependencies
            })
        )
}
