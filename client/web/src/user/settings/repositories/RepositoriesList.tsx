import React from 'react'

import { useHistory, useLocation } from 'react-router'
import { Observable } from 'rxjs'

import { ErrorLike } from '@sourcegraph/common'
import { Container } from '@sourcegraph/wildcard'

import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
    Connection,
} from '../../../components/FilteredConnection'
import { SiteAdminRepositoryFields, UserRepositoriesResult, Maybe } from '../../../graphql-operations'

import { RepositoryNode } from './RepositoryNode'

import styles from './RepositoriesList.module.scss'

interface Props {
    queryRepos: (
        args: FilteredConnectionQueryArguments
    ) => Observable<{
        __typename?: 'RepositoryConnection' | undefined
        totalCount: Maybe<number>
        nodes: ({
            __typename?: 'Repository' | undefined
        } & SiteAdminRepositoryFields)[]
        pageInfo: {
            __typename?: 'PageInfo' | undefined
            hasNextPage: boolean
        }
    }>
    updateReposList: boolean
    onRepoQueryUpdate?: (value: Connection<SiteAdminRepositoryFields> | ErrorLike | undefined, query: string) => void
    repoFilters: FilteredConnectionFilter[]
}

interface RowProps {
    node: SiteAdminRepositoryFields
}

const Row: React.FunctionComponent<React.PropsWithChildren<RowProps>> = props => (
    <RepositoryNode
        name={props.node.name}
        url={props.node.url}
        serviceType={props.node.externalRepository.serviceType.toUpperCase()}
        mirrorInfo={props.node.mirrorInfo}
        isPrivate={props.node.isPrivate}
    />
)

const NoMatchedRepos = (
    <div className="border rounded p-3">
        <small>No repositories matched.</small>
    </div>
)

const TotalCountSummary: React.FunctionComponent<React.PropsWithChildren<{ totalCount: number }>> = ({
    totalCount,
}) => (
    <div className="d-inline-block mt-4 mr-2">
        <small>
            {totalCount} {totalCount === 1 ? 'repository' : 'repositories'} total
        </small>
    </div>
)

/**
 * A page displaying the repositories for this user.
 */
export const RepositoriesList: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    queryRepos,
    updateReposList,
    onRepoQueryUpdate,
    repoFilters,
}) => {
    const location = useLocation()
    const history = useHistory()

    return (
        <Container>
            <FilteredConnection<SiteAdminRepositoryFields, Omit<UserRepositoriesResult, 'node'>>
                className="table mb-0"
                defaultFirst={15}
                compact={false}
                noun="repository"
                pluralNoun="repositories"
                queryConnection={queryRepos}
                updateOnChange={String(updateReposList)}
                nodeComponent={Row}
                listComponent="table"
                listClassName="w-100"
                onUpdate={onRepoQueryUpdate}
                filters={repoFilters}
                history={history}
                location={location}
                emptyElement={NoMatchedRepos}
                totalCountSummaryComponent={TotalCountSummary}
                inputClassName={styles.filterInput}
            />
        </Container>
    )
}

export const defaultFilters: FilteredConnectionFilter[] = [
    {
        label: 'Status',
        type: 'select',
        id: 'status',
        tooltip: 'Repository status',
        values: [
            {
                value: 'all',
                label: 'All',
                args: {},
            },
            {
                value: 'cloned',
                label: 'Cloned',
                args: { cloned: true, notCloned: false },
            },
            {
                value: 'not-cloned',
                label: 'Not Cloned',
                args: { cloned: false, notCloned: true },
            },
        ],
    },
    {
        label: 'Code host',
        type: 'select',
        id: 'code-host',
        tooltip: 'Code host',
        values: [
            {
                value: 'all',
                label: 'All',
                args: {},
            },
        ],
    },
]
