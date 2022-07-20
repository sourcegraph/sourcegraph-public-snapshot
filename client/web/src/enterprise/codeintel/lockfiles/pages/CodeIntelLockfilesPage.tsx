/* eslint-disable arrow-body-style */
import { FunctionComponent, useCallback, useEffect } from 'react'

import { useApolloClient } from '@apollo/client'
import classNames from 'classnames'
import { RouteComponentProps } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '../../../../components/FilteredConnection'
import { PageTitle } from '../../../../components/PageTitle'
import { LockfileIndexFields } from '../../../../graphql-operations'
import { CodeIntelLockfileNode, CodeIntelLockfileNodeProps } from '../components/CodeIntelLockfileIndexNode'
import { EmptyLockfiles } from '../components/EmptyLockfiles'
import { queryLockfileIndexesList as defaultQueryLockfileIndexesList } from '../hooks/queryLockfileIndexesList'

import styles from './CodeIntelLockfilesPage.module.scss'

export interface CodeIntelLockfilesPageProps extends RouteComponentProps<{}>, TelemetryProps {
    queryLockfileIndexesList?: typeof defaultQueryLockfileIndexesList
}

const filters: FilteredConnectionFilter[] = []

export const CodeIntelLockfilesPage: FunctionComponent<React.PropsWithChildren<CodeIntelLockfilesPageProps>> = ({
    queryLockfileIndexesList = defaultQueryLockfileIndexesList,
    telemetryService,
    history,
    ...props
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelLockfiles'), [telemetryService])

    const apolloClient = useApolloClient()
    const queryLockfileIndexes = useCallback(
        (args: FilteredConnectionQueryArguments) => {
            return queryLockfileIndexesList({ ...args }, apolloClient)
        },
        [queryLockfileIndexesList, apolloClient]
    )

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
                    <FilteredConnection<LockfileIndexFields, Omit<CodeIntelLockfileNodeProps, 'node'>>
                        listComponent="div"
                        listClassName={classNames(styles.grid, 'mb-3')}
                        inputClassName="w-auto"
                        noun="lockfile index"
                        pluralNoun="lockfile indexes"
                        nodeComponent={CodeIntelLockfileNode}
                        queryConnection={queryLockfileIndexes}
                        history={history}
                        location={props.location}
                        cursorPaging={true}
                        filters={filters}
                        hideSearch={true}
                        emptyElement={<EmptyLockfiles />}
                    />
                </div>
            </Container>
        </div>
    )
}
