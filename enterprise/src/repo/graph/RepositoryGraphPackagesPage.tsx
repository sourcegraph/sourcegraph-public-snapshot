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

interface PackageNodePropsCommon {
    repo: GQL.IRepository
    rev: string | undefined
    commitID: string
}

interface PackageNodeProps extends PackageNodePropsCommon {
    node: GQL.IPackage
}

interface PackageNodeState {
    internalReferenceCountOrError?: GQL.IApproximateCount | null | ErrorLike
    externalReferenceCountOrError?: GQL.IApproximateCount | null | ErrorLike
}

class PackageNode extends React.PureComponent<PackageNodeProps, PackageNodeState> {
    public state: PackageNodeState = {}

    private componentUpdates = new Subject<PackageNodeProps>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    map(({ node }) => node),
                    distinctUntilChanged(),
                    filter(node => !!node.internalReferences),
                    switchMap(node =>
                        this.fetchInternalReferenceCount(node).pipe(
                            catchError(error => {
                                console.error(error)
                                this.setState({ internalReferenceCountOrError: error })
                                return []
                            })
                        )
                    )
                )
                .subscribe(
                    references => this.setState({ internalReferenceCountOrError: references }),
                    error => console.error(error)
                )
        )

        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    map(({ node }) => node),
                    distinctUntilChanged(),
                    filter(node => !!node.externalReferences),
                    switchMap(node =>
                        this.fetchExternalReferenceCount(node).pipe(
                            catchError(error => {
                                console.error(error)
                                this.setState({ externalReferenceCountOrError: error })
                                return []
                            })
                        )
                    )
                )
                .subscribe(
                    references => this.setState({ externalReferenceCountOrError: references }),
                    error => console.error(error)
                )
        )

        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(nextProps: PackageNodeProps): void {
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
                        .filter(({ key, value }) => !!value && key !== 'doc')
                        .map(({ key, value }) => (
                            <span key={key} className="repository-graph-page__node-datum">
                                <span className="repository-graph-page__node-datum-name">{key}:</span>
                                <span className="repository-graph-page__node-datum-value">{String(value)}</span>
                            </span>
                        ))}
                </div>
                <div className="repository-graph-page__node-actions">
                    {node.internalReferences ? (
                        <>
                            {this.state.internalReferenceCountOrError === undefined && (
                                <LoadingSpinner className="icon-inline" />
                            )}
                            {isErrorLike(this.state.internalReferenceCountOrError) ? (
                                <span title={this.state.internalReferenceCountOrError.message}>
                                    <AlertCircleIcon className="icon-inline" />
                                </span>
                            ) : (
                                <Link
                                    to={this.urlToInternalSearch(node.internalReferences.queryString)}
                                    className="btn btn-link btn-sm"
                                >
                                    {this.state.internalReferenceCountOrError === undefined ||
                                    this.state.internalReferenceCountOrError === null ? (
                                        'References in this repository'
                                    ) : (
                                        <>
                                            {this.state.internalReferenceCountOrError.label}{' '}
                                            {pluralize('reference', this.state.internalReferenceCountOrError.count)} in
                                            this repository
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
                    {node.externalReferences ? (
                        <>
                            {this.state.externalReferenceCountOrError === undefined &&
                                this.state.internalReferenceCountOrError !== undefined && (
                                    <LoadingSpinner className="icon-inline" />
                                )}
                            {isErrorLike(this.state.externalReferenceCountOrError) ? (
                                !isErrorLike(this.state.internalReferenceCountOrError) && (
                                    <span title={this.state.externalReferenceCountOrError.message}>
                                        <AlertCircleIcon className="icon-inline" />
                                    </span>
                                )
                            ) : (
                                <Link
                                    to={this.urlToExternalSearch(node.externalReferences.queryString)}
                                    className="btn btn-link btn-sm"
                                >
                                    {this.state.externalReferenceCountOrError === undefined ||
                                    this.state.externalReferenceCountOrError === null ? (
                                        'External dependents'
                                    ) : (
                                        <>
                                            {this.state.externalReferenceCountOrError.label}{' '}
                                            {pluralize(
                                                'external dependent',
                                                this.state.externalReferenceCountOrError.count
                                            )}
                                        </>
                                    )}
                                </Link>
                            )}
                        </>
                    ) : (
                        !!node.internalReferences && (
                            <DotsHorizontalIcon
                                className="icon-inline repository-graph-page__node-dotdotdot"
                                data-tooltip={`External reference search is not supported by the language server (${
                                    this.props.node.language
                                }).`}
                            />
                        )
                    )}
                </div>
            </li>
        )
    }

    private urlToInternalSearch(partialQuery: string | null): string {
        if (partialQuery === null) {
            return ''
        }
        const query = `${searchQueryForRepoRev(this.props.repo.name, this.props.rev)}type:ref ${partialQuery}`
        return `/search?${buildSearchURLQuery({ query })}`
    }

    private urlToExternalSearch(partialQuery: string | null): string {
        if (partialQuery === null) {
            return ''
        }
        return `/search?${buildSearchURLQuery({ query: `type:ref ${partialQuery}` })}`
    }

    private fetchInternalReferenceCount = (node: GQL.IPackage): Observable<GQL.IApproximateCount | null> =>
        queryGraphQL(
            gql`
                query PackageInternalReferenceCount($pkg: ID!) {
                    node(id: $pkg) {
                        ... on Package {
                            internalReferences {
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
            { pkg: node.id }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.node) {
                    throw createAggregateError(errors)
                }
                const pkg = data.node as GQL.IPackage
                if (!pkg.internalReferences) {
                    throw createAggregateError(errors)
                }
                return pkg.internalReferences.approximateCount
            })
        )

    private fetchExternalReferenceCount = (node: GQL.IPackage): Observable<GQL.IApproximateCount | null> =>
        queryGraphQL(
            gql`
                query PackageExternalReferenceCount($pkg: ID!) {
                    node(id: $pkg) {
                        ... on Package {
                            externalReferences {
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
            { pkg: node.id }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.node) {
                    throw createAggregateError(errors)
                }
                const pkg = data.node as GQL.IPackage
                if (!pkg.externalReferences) {
                    throw createAggregateError(errors)
                }
                return pkg.externalReferences.approximateCount
            })
        )
}

interface Props extends RouteComponentProps<{}> {
    repo: GQL.IRepository
    rev: string | undefined
    commitID: string
}

interface State {}

class FilteredPackageConnection extends FilteredConnection<GQL.IPackage, PackageNodePropsCommon> {}

/**
 * The repository graph packages page.
 */
export class RepositoryGraphPackagesPage extends React.PureComponent<Props, State> {
    public state: State = {}

    private repoChanges = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('RepoGraphPackages')
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
                <PageTitle title="Packages - Repository graph" />
                <h2>Packages</h2>
                <FilteredPackageConnection
                    noun="package"
                    pluralNoun="packages"
                    queryConnection={this.fetchPackages}
                    nodeComponent={PackageNode}
                    nodeComponentProps={
                        {
                            repo: this.props.repo,
                            rev: this.props.rev,
                            commitID: this.props.commitID,
                        } as PackageNodePropsCommon
                    }
                    defaultFirst={10}
                    updates={this.repoChanges}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private fetchPackages = (args: { first?: number; query?: string }): Observable<GQL.IPackageConnection> =>
        queryGraphQL(
            gql`
                query RepositoryPackages($repo: ID!, $commitID: String!, $first: Int, $query: String) {
                    node(id: $repo) {
                        ... on Repository {
                            commit(rev: $commitID) {
                                packages(first: $first, query: $query) {
                                    nodes {
                                        id
                                        language
                                        data {
                                            key
                                            value
                                        }
                                        internalReferences {
                                            queryString
                                        }
                                        externalReferences {
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
                if (!repo.commit || !repo.commit.packages || !repo.commit.packages.nodes) {
                    throw createAggregateError(errors)
                }
                return repo.commit.packages
            })
        )
}
