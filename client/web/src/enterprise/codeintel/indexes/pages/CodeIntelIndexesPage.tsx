import { FunctionComponent, useCallback, useEffect, useMemo } from 'react'

import { useApolloClient } from '@apollo/client'
import classNames from 'classnames'
import { RouteComponentProps } from 'react-router'
import { Subject } from 'rxjs'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../auth'
import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '../../../../components/FilteredConnection'
import { PageTitle } from '../../../../components/PageTitle'
import { LsifIndexFields, LSIFIndexState } from '../../../../graphql-operations'
import { FlashMessage } from '../../configuration/components/FlashMessage'
import { CodeIntelIndexNode, CodeIntelIndexNodeProps } from '../components/CodeIntelIndexNode'
import { EmptyAutoIndex } from '../components/EmptyAutoIndex'
import { EnqueueForm } from '../components/EnqueueForm'
import { queryLsifIndexList as defaultQueryLsifIndexList } from '../hooks/queryLsifIndexList'
import { queryLsifIndexListByRepository as defaultQueryLsifIndexListByRepository } from '../hooks/queryLsifIndexListByRepository'

import styles from './CodeIntelIndexesPage.module.scss'

export interface CodeIntelIndexesPageProps extends RouteComponentProps<{}>, TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    repo?: { id: string }
    queryLsifIndexListByRepository?: typeof defaultQueryLsifIndexListByRepository
    queryLsifIndexList?: typeof defaultQueryLsifIndexList
    now?: () => Date
}

const filters: FilteredConnectionFilter[] = [
    {
        id: 'filters',
        label: 'Index state',
        type: 'select',
        values: [
            {
                label: 'All',
                value: 'all',
                tooltip: 'Show all indexes',
                args: {},
            },
            {
                label: 'Completed',
                value: 'completed',
                tooltip: 'Show completed indexes only',
                args: { state: LSIFIndexState.COMPLETED },
            },
            {
                label: 'Errored',
                value: 'errored',
                tooltip: 'Show errored indexes only',
                args: { state: LSIFIndexState.ERRORED },
            },
            {
                label: 'Processing',
                value: 'processing',
                tooltip: 'Show processing indexes only',
                args: { state: LSIFIndexState.PROCESSING },
            },
            {
                label: 'Queued',
                value: 'queued',
                tooltip: 'Show queued indexes only',
                args: { state: LSIFIndexState.QUEUED },
            },
        ],
    },
]

export const CodeIntelIndexesPage: FunctionComponent<React.PropsWithChildren<CodeIntelIndexesPageProps>> = ({
    authenticatedUser,
    repo,
    queryLsifIndexListByRepository = defaultQueryLsifIndexListByRepository,
    queryLsifIndexList = defaultQueryLsifIndexList,
    now,
    telemetryService,
    history,
    ...props
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelIndexes'), [telemetryService])

    const apolloClient = useApolloClient()
    const queryIndexes = useCallback(
        (args: FilteredConnectionQueryArguments) => {
            if (repo?.id) {
                return queryLsifIndexListByRepository(args, repo?.id, apolloClient)
            }

            return queryLsifIndexList(args, apolloClient)
        },
        [repo?.id, queryLsifIndexListByRepository, queryLsifIndexList, apolloClient]
    )

    const querySubject = useMemo(() => new Subject<string>(), [])

    return (
        <div className="code-intel-indexes">
            <PageTitle title="Auto-indexing jobs" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Auto-indexing jobs' }]}
                description={`Auto-indexing jobs ${repo ? 'for this repository' : 'over all repositories'}.`}
                className="mb-3"
            />

            {repo && authenticatedUser?.siteAdmin && (
                <Container className="mb-2">
                    <EnqueueForm repoId={repo.id} querySubject={querySubject} />
                </Container>
            )}

            {history.location.state && (
                <FlashMessage state={history.location.state.modal} message={history.location.state.message} />
            )}

            <Container>
                <div className="list-group position-relative">
                    <FilteredConnection<LsifIndexFields, Omit<CodeIntelIndexNodeProps, 'node'>>
                        listComponent="div"
                        inputClassName="w-auto"
                        listClassName={classNames(styles.grid, 'mb-3')}
                        noun="index"
                        pluralNoun="indexes"
                        querySubject={querySubject}
                        nodeComponent={CodeIntelIndexNode}
                        nodeComponentProps={{ now }}
                        queryConnection={queryIndexes}
                        history={history}
                        location={props.location}
                        cursorPaging={true}
                        filters={filters}
                        emptyElement={<EmptyAutoIndex />}
                    />
                </div>
            </Container>
        </div>
    )
}
