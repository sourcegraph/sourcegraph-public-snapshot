import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../components/FilteredConnection'
import { displayRepoPath } from '../components/RepoFileLink'
import { eventLogger } from '../tracking/eventLogger'
import { createAggregateError } from '../util/errors'
import { CircleChevronLeftIcon } from '../util/icons'

function fetchRepositories(args: { first?: number; query?: string }): Observable<GQL.IRepositoryConnection> {
    return queryGraphQL(
        gql`
            query Repositories($first: Int, $query: String) {
                repositories(first: $first, query: $query) {
                    nodes {
                        id
                        name
                    }
                    totalCount
                    pageInfo {
                        hasNextPage
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.repositories) {
                throw createAggregateError(errors)
            }
            return data.repositories
        })
    )
}

interface RepositoryNodeProps {
    node: GQL.IRepository
    currentRepo?: GQL.ID
}

const RepositoryNode: React.SFC<RepositoryNodeProps> = ({ node, currentRepo }) => (
    <li key={node.id} className="connection-popover__node">
        <Link
            to={`/${node.name}`}
            className={`connection-popover__node-link ${
                node.id === currentRepo ? 'connection-popover__node-link--active' : ''
            }`}
        >
            {displayRepoPath(node.name)}
            {node.id === currentRepo && (
                <CircleChevronLeftIcon className="icon-inline connection-popover__node-link-icon" />
            )}
        </Link>
    </li>
)

interface Props {
    /**
     * The current repository (shown as selected in the list), if any.
     */
    currentRepo?: GQL.ID

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
            <div className="repositories-popover connection-popover">
                <FilteredRepositoryConnection
                    className="connection-popover__content"
                    showMoreClassName="connection-popover__show-more"
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
