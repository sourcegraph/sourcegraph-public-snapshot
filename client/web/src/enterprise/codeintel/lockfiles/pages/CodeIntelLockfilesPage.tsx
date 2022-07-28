import { FunctionComponent, useEffect } from 'react'

import { RouteComponentProps } from 'react-router'

import { createAggregateError } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { useConnection } from '../../../../components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ConnectionError,
    SummaryContainer,
    ShowMoreButton,
} from '../../../../components/FilteredConnection/ui'
import { PageTitle } from '../../../../components/PageTitle'
import { LockfileIndexesResult, LockfileIndexesVariables, LockfileIndexFields } from '../../../../graphql-operations'
import { CodeIntelLockfileNode } from '../components/CodeIntelLockfileIndexNode'

import { LOCKFILE_INDEXES_LIST } from './queries'

import styles from './CodeIntelLockfilesPage.module.scss'

export interface CodeIntelLockfilesPageProps extends RouteComponentProps<{}>, TelemetryProps {}

const DEFAULT_LOCKFILE_INDEXES_PAGE_SIZE = 50

export const CodeIntelLockfilesPage: FunctionComponent<React.PropsWithChildren<CodeIntelLockfilesPageProps>> = ({
    telemetryService,
}) => {
    useEffect(() => telemetryService.logPageView('CodeIntelLockfiles'), [telemetryService])

    const { connection, loading, error, hasNextPage, fetchMore } = useConnection<
        LockfileIndexesResult,
        LockfileIndexesVariables,
        LockfileIndexFields
    >({
        query: LOCKFILE_INDEXES_LIST,
        variables: { first: DEFAULT_LOCKFILE_INDEXES_PAGE_SIZE, after: null },
        getConnection: ({ data, errors }) => {
            if (!data || !data.lockfileIndexes) {
                throw createAggregateError(errors)
            }
            return data.lockfileIndexes
        },
        options: {
            fetchPolicy: 'cache-and-network',
        },
    })

    return (
        <div className="code-intel-lockfiles">
            <PageTitle title="Lockfile indexes" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Lockfile indexes' }]}
                description="Lockfile indexes created by lockfile-indexing"
                className="mb-3"
            />

            <Container>
                <div className="list-group position-relative">
                    <ConnectionContainer>
                        {error && <ConnectionError errors={[error.message]} />}
                        {connection && (
                            <ConnectionList className={styles.grid}>
                                {connection.nodes.map(node => (
                                    <CodeIntelLockfileNode key={node.id} node={node} />
                                ))}
                            </ConnectionList>
                        )}
                        {loading && <ConnectionLoading />}
                        {!loading && connection && (
                            <SummaryContainer>
                                <ConnectionSummary
                                    connection={connection}
                                    first={DEFAULT_LOCKFILE_INDEXES_PAGE_SIZE}
                                    noun="lockfile index"
                                    pluralNoun="lockfile indexes"
                                    hasNextPage={hasNextPage}
                                    noSummaryIfAllNodesVisible={true}
                                    compact={true}
                                />
                                {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                            </SummaryContainer>
                        )}
                    </ConnectionContainer>
                </div>
            </Container>
        </div>
    )
}
