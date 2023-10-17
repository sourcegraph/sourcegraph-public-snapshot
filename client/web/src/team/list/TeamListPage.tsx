import React, { useState } from 'react'

import { mdiAccountMultiple, mdiPlus } from '@mdi/js'
import classNames from 'classnames'

import { Button, Link, Icon, PageHeader, Container, useDebounce, ProductStatusBadge } from '@sourcegraph/wildcard'

import type { UseShowMorePaginationResult } from '../../components/FilteredConnection/hooks/useShowMorePagination'
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
import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import type { ListTeamFields, ListTeamsOfParentResult, ListTeamsResult } from '../../graphql-operations'

import { useChildTeams, useTeams } from './backend'
import { TeamNode } from './TeamNode'

export interface TeamListPageProps {}

/**
 * A page displaying the teams on this site.
 */
export const TeamListPage: React.FunctionComponent<React.PropsWithChildren<TeamListPageProps>> = () => {
    const [searchValue, setSearchValue] = useState('')
    const query = useDebounce(searchValue, 200)

    const connection = useTeams(query)

    return (
        <Page className="mb-3">
            <PageTitle title="Teams" />
            <PageHeader
                actions={
                    <>
                        <Button to="/teams/new" variant="primary" as={Link}>
                            <Icon aria-hidden={true} svgPath={mdiPlus} /> Create team
                        </Button>
                    </>
                }
                description={
                    <>
                        A team is a set of users. See the <Link to="/help/admin/teams">Teams documentation</Link> for
                        more information about configuring teams.
                    </>
                }
                className="mb-3"
            >
                <PageHeader.Heading as="h2" styleAs="h1">
                    <PageHeader.Breadcrumb icon={mdiAccountMultiple}>
                        Teams <ProductStatusBadge status="experimental" />
                    </PageHeader.Breadcrumb>
                </PageHeader.Heading>
            </PageHeader>

            <Container className="mb-3">
                <TeamList searchValue={searchValue} setSearchValue={setSearchValue} query={query} {...connection} />
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
    const [searchValue, setSearchValue] = useState('')
    const query = useDebounce(searchValue, 200)

    const connection = useChildTeams(parentTeam, query)

    return (
        <>
            <div className="d-flex justify-content-end mb-3">
                <Button to={`/teams/new?parentTeam=${parentTeam}`} variant="primary" as={Link}>
                    <Icon aria-hidden={true} svgPath={mdiPlus} /> Create child team
                </Button>
            </div>
            <Container className="mb-3">
                <TeamList searchValue={searchValue} setSearchValue={setSearchValue} query={query} {...connection} />
            </Container>
        </>
    )
}

interface TeamListProps extends UseShowMorePaginationResult<ListTeamsResult | ListTeamsOfParentResult, ListTeamFields> {
    searchValue: string
    setSearchValue: (value: string) => void
    query: string
    className?: string
}

export const TeamList: React.FunctionComponent<TeamListProps> = ({
    fetchMore,
    hasNextPage,
    loading,
    refetchAll,
    connection,
    error,
    searchValue,
    setSearchValue,
    query,
    className,
}) => (
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
