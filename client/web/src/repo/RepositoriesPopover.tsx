import classNames from 'classnames'
import React, { useCallback, useEffect } from 'react'
import { Link, useHistory, useLocation } from 'react-router-dom'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { gql } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { createAggregateError } from '@sourcegraph/shared/src/util/errors'

import { queryGraphQL } from '../backend/graphql'
import { FilteredConnection, FilteredConnectionQueryArguments } from '../components/FilteredConnection'
import { eventLogger } from '../tracking/eventLogger'

function fetchRepositories(args: { first?: number; query?: string }): Observable<GQL.IRepositoryConnection> {
    return queryGraphQL(
        gql`
            query RepositoriesForPopover($first: Int, $query: String) {
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
    currentRepo?: Scalars['ID']
}

const RepositoryNode: React.FunctionComponent<RepositoryNodeProps> = ({ node, currentRepo }) => (
    <li key={node.id} className="connection-popover__node">
        <Link
            to={`/${node.name}`}
            className={classNames(
                'connection-popover__node-link',
                node.id === currentRepo && 'connection-popover__node-link--active'
            )}
        >
            {displayRepoName(node.name)}
        </Link>
    </li>
)

interface RepositoriesPopoverProps {
    /**
     * The current repository (shown as selected in the list), if any.
     */
    currentRepo?: Scalars['ID']
}

class FilteredRepositoryConnection extends FilteredConnection<GQL.IRepository> {}

/**
 * A popover that displays a searchable list of repositories.
 */
export const RepositoriesPopover: React.FunctionComponent<RepositoriesPopoverProps> = ({ currentRepo }) => {
    const location = useLocation()
    const history = useHistory()

    useEffect(() => {
        eventLogger.logViewEvent('RepositoriesPopover')
    }, [])

    const queryRepositories = useCallback(
        (args: FilteredConnectionQueryArguments): Observable<GQL.IRepositoryConnection> =>
            fetchRepositories({ ...args }),
        []
    )

    const nodeProps: Pick<RepositoryNodeProps, 'currentRepo'> = { currentRepo }

    return (
        <div className="repositories-popover connection-popover">
            <FilteredRepositoryConnection
                className="connection-popover__content"
                inputClassName="connection-popover__input"
                listClassName="connection-popover__nodes"
                compact={true}
                noun="repository"
                pluralNoun="repositories"
                queryConnection={queryRepositories}
                nodeComponent={RepositoryNode}
                nodeComponentProps={nodeProps}
                defaultFirst={10}
                autoFocus={true}
                history={history}
                location={location}
                noSummaryIfAllNodesVisible={true}
                useURLQuery={false}
            />
        </div>
    )
}
