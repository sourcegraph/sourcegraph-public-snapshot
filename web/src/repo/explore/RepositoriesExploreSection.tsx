import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import { Observable, Subscription } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { RepoLink } from '../../../../shared/src/components/RepoLink'
import { gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { pluralize } from '../../../../shared/src/util/strings'
import { buildSearchURLQuery } from '../../../../shared/src/util/url'
import { queryGraphQL } from '../../backend/graphql'
import { PatternTypeProps } from '../../search'
import { ErrorAlert } from '../../components/alerts'

const LOADING: 'loading' = 'loading'

interface State {
    /** The repositories, loading, or an error. */
    repositoriesOrError: typeof LOADING | GQL.IRepositoryConnection | ErrorLike
}

/**
 * An explore section that shows a few repositories and a link to all.
 */
export class RepositoriesExploreSection extends React.PureComponent<Omit<PatternTypeProps, 'setPatternType'>, State> {
    private static QUERY_REPOSITORIES_ARGS: { first: number } & Pick<GQL.IRepositoriesOnQueryArguments, 'names'> = {
        // Show sample repositories on Sourcegraph.com.
        names: window.context.sourcegraphDotComMode
            ? [
                  'github.com/sourcegraph/sourcegraph',
                  'github.com/theupdateframework/notary',
                  'github.com/pallets/flask',
                  'github.com/ReactiveX/rxjs',
              ]
            : null,
        first: 4,
    }

    public state: State = { repositoriesOrError: LOADING }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            queryRepositories(RepositoriesExploreSection.QUERY_REPOSITORIES_ARGS)
                .pipe(catchError(err => [asError(err)]))
                .subscribe(repositoriesOrError => this.setState({ repositoriesOrError }))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const repositoriesOrError: (typeof LOADING | GQL.IRepository)[] | ErrorLike =
            this.state.repositoriesOrError === LOADING
                ? Array(RepositoriesExploreSection.QUERY_REPOSITORIES_ARGS.first).fill(LOADING)
                : isErrorLike(this.state.repositoriesOrError)
                ? this.state.repositoriesOrError
                : this.state.repositoriesOrError.nodes

        const itemClass = 'py-2'

        // Only show total count if it is counting *all* repositories (i.e., no filter args are specified).
        const queryingAllRepositories = RepositoriesExploreSection.QUERY_REPOSITORIES_ARGS.names === null
        const totalCount =
            queryingAllRepositories &&
            this.state.repositoriesOrError !== LOADING &&
            !isErrorLike(this.state.repositoriesOrError) &&
            typeof this.state.repositoriesOrError.totalCount === 'number'
                ? this.state.repositoriesOrError.totalCount
                : undefined

        return (
            <div className="card">
                <h3 className="card-header">Repositories</h3>
                <div className="list-group list-group-flush">
                    {isErrorLike(repositoriesOrError) ? (
                        <ErrorAlert error={repositoriesOrError} />
                    ) : repositoriesOrError.length === 0 ? (
                        <p>No repositories.</p>
                    ) : (
                        repositoriesOrError.map((repo /* or loading */, i) =>
                            repo === LOADING ? (
                                <div key={i} className={`${itemClass} list-group-item`}>
                                    <h4 className="text-muted mb-0">⋯</h4>&nbsp;
                                </div>
                            ) : (
                                <Link
                                    key={i}
                                    className={`${itemClass} list-group-item list-group-item-action text-truncate`}
                                    to={repo.url}
                                >
                                    <h4 className="mb-0 text-truncate">
                                        <RepoLink to={null} repoName={repo.name} />
                                    </h4>
                                    <span>{repo.description || <>&nbsp;</>}</span>
                                </Link>
                            )
                        )
                    )}
                </div>
                {typeof totalCount === 'number' && totalCount > 0 && (
                    <div className="card-footer">
                        <Link to={`/search?${buildSearchURLQuery('repo:', this.props.patternType)}`}>
                            View all {totalCount} {pluralize('repository', totalCount, 'repositories')}
                            <ChevronRightIcon className="icon-inline" />
                        </Link>
                    </div>
                )}
            </div>
        )
    }
}

function queryRepositories(
    args: Pick<GQL.IRepositoriesOnQueryArguments, 'first' | 'names'>
): Observable<GQL.IRepositoryConnection> {
    return queryGraphQL(
        gql`
            query ExploreRepositories($first: Int, $names: [String!]) {
                repositories(first: $first, names: $names) {
                    nodes {
                        name
                        description
                        url
                    }
                    totalCount(precise: false)
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.repositories || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            return data.repositories
        })
    )
}
