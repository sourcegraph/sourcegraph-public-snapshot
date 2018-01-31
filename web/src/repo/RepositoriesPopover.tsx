import CircleChevronLeft from '@sourcegraph/icons/lib/CircleChevronLeft'
import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { gql, queryGraphQL } from '../backend/graphql'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../components/FilteredConnection'
import { displayRepoPath } from '../components/RepoFileLink'
import { eventLogger } from '../tracking/eventLogger'
import { createAggregateError } from '../util/errors'

function fetchRepositories(args: { first?: number; query?: string }): Observable<GQL.IRepositoryConnection> {
    return queryGraphQL(
        gql`
            query Repositories($first: Int, $query: String) {
                site {
                    repositories(first: $first, query: $query) {
                        nodes {
                            id
                            uri
                        }
                        totalCount
                        pageInfo {
                            hasNextPage
                        }
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.site || !data.site.repositories) {
                throw createAggregateError(errors)
            }
            return data.site.repositories
        })
    )
}

interface RepositoryNodeProps {
    node: GQL.IRepository
    currentRepo?: GQLID
}

export const RepositoryNode: React.SFC<RepositoryNodeProps> = ({ node, currentRepo }) => (
    <li key={node.id} className="popover__node">
        <Link
            to={`/${node.uri}`}
            className={`popover__node-link ${node.id === currentRepo ? 'popover__node-link--active' : ''}`}
        >
            {displayRepoPath(node.uri)}
            {node.id === currentRepo && <CircleChevronLeft className="icon-inline popover__node-link-icon" />}
        </Link>
    </li>
)

interface Props {
    /**
     * The current repository (shown as selected in the list), if any.
     */
    currentRepo?: GQLID

    history: H.History
    location: H.Location
}

class FilteredRepositoryConnection extends FilteredConnection<GQL.IRepository> {}

/**
 * A popover that displays a searchable list of repositories.
 */
export class RepositoriesPopover extends React.PureComponent<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('RepositoriesPopover')
    }

    public render(): JSX.Element | null {
        const nodeProps: Pick<RepositoryNodeProps, 'currentRepo'> = { currentRepo: this.props.currentRepo }

        return (
            <div className="repositories-popover popover">
                <FilteredRepositoryConnection
                    className="popover__content"
                    compact={true}
                    noun="repository"
                    pluralNoun="repositories"
                    queryConnection={this.queryRepositories}
                    nodeComponent={RepositoryNode}
                    nodeComponentProps={nodeProps}
                    defaultFirst={10}
                    autoFocus={true}
                    history={this.props.history}
                    location={this.props.location}
                    noSummaryIfAllNodesVisible={true}
                    shouldUpdateURLQuery={false}
                />
            </div>
        )
    }

    private queryRepositories = (args: FilteredConnectionQueryArgs) => fetchRepositories({ ...args })
}
