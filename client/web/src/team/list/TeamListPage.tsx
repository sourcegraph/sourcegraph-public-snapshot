import React, { useState } from 'react'

import { mdiPlus } from '@mdi/js'

import { Button, Link, Icon, PageHeader, Container, useDebounce } from '@sourcegraph/wildcard'

import { useChildTeams, useTeams } from './backend'
import { Page } from '../../components/Page'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionLoading,
    ConnectionList,
    SummaryContainer,
    ConnectionSummary,
    ShowMoreButton,
    ConnectionForm,
} from '../../components/FilteredConnection/ui'
import { PageTitle } from '../../components/PageTitle'
import { TeamNode } from './TeamNode'
import { UseShowMorePaginationResult } from '../../components/FilteredConnection/hooks/useShowMorePagination'
import { ListTeamFields, ListTeamsOfParentResult, ListTeamsResult } from '../../graphql-operations'
import classNames from 'classnames'

export interface TeamListPageProps {}

/**
 * A page displaying the teams on this site.
 */
export const TeamListPage: React.FunctionComponent<React.PropsWithChildren<TeamListPageProps>> = () => {
    const connection = useTeams()

    return (
        <Page className="mb-3">
            <PageTitle title="Teams" />
            <PageHeader
                path={[{ text: 'Teams' }]}
                actions={
                    <>
                        <Button to="/teams/new" variant="primary" as={Link}>
                            <Icon aria-hidden={true} svgPath={mdiPlus} /> Create team
                        </Button>
                    </>
                }
                description={
                    <>
                        A team is a set of users. See <Link to="/help/teams">Teams documentation</Link> for more
                        information about configuring teams.
                    </>
                }
                className="mb-3"
            />

            <Container className="mb-3">
                <TeamList connectionFunction={connection} />
            </Container>
        </Page>
    )
}

export interface ChildTeamListPageProps {
    parentTeam: string
}

/**
 * A page displaying the child teams of a given teams.
 */
export const ChildTeamListPage: React.FunctionComponent<React.PropsWithChildren<ChildTeamListPageProps>> = ({
    parentTeam,
}) => {
    const connection = useChildTeams(parentTeam)

    return (
        <>
            <div className="d-flex justify-content-end mb-3">
                <Button to={`/teams/new?parentTeam=${parentTeam}`} variant="primary" as={Link}>
                    <Icon aria-hidden={true} svgPath={mdiPlus} /> Create child team
                </Button>
            </div>
            <Container className="mb-3">
                <TeamList connectionFunction={connection} />
            </Container>
        </>
    )
}

export const TeamList: React.FunctionComponent<{
    connectionFunction: (
        search: string | null
    ) => UseShowMorePaginationResult<ListTeamsResult | ListTeamsOfParentResult, ListTeamFields>
    className?: string
}> = ({ connectionFunction, className }) => {
    const [searchValue, setSearchValue] = useState('')
    const query = useDebounce(searchValue, 200)

    const { fetchMore, hasNextPage, loading, refetchAll, connection, error } = connectionFunction(query)

    return (
        <ConnectionContainer className={classNames(className)}>
            <ConnectionForm
                inputValue={searchValue}
                onInputChange={event => setSearchValue(event.target.value)}
                inputPlaceholder="Search teams"
            />

            {error && <ConnectionError errors={[error.message]} />}
            {loading && !connection && <ConnectionLoading />}
            <ConnectionList as="ul" className="list-group" aria-label="Teams">
                {connection?.nodes?.map(node => (
                    <TeamNode key={node.id} node={node} refetchAll={refetchAll} />
                ))}
            </ConnectionList>
            {connection && (
                <SummaryContainer className="mt-2">
                    <ConnectionSummary
                        first={15}
                        centered={true}
                        connection={connection}
                        noun="team"
                        pluralNoun="teams"
                        hasNextPage={hasNextPage}
                        connectionQuery={query}
                    />
                    {hasNextPage && <ShowMoreButton centered={true} onClick={fetchMore} />}
                </SummaryContainer>
            )}
        </ConnectionContainer>
    )
}
