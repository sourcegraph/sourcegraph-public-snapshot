import React, { useCallback, useMemo } from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable, Subject } from 'rxjs'

import { Container, PageHeader } from '@sourcegraph/wildcard'

import { FilteredConnection } from '../../../components/FilteredConnection'
import {
    BatchChangesCodeHostFields,
    BatchChangesCodeHostsFields,
    Scalars,
    UserBatchChangesCodeHostsVariables,
} from '../../../graphql-operations'

import {
    queryUserBatchChangesCodeHosts as _queryUserBatchChangesCodeHosts,
    queryGlobalBatchChangesCodeHosts as _queryGlobalBatchChangesCodeHosts,
} from './backend'
import { CodeHostConnectionNode, CodeHostConnectionNodeProps } from './CodeHostConnectionNode'

export interface CodeHostConnectionsProps extends Pick<RouteComponentProps, 'history' | 'location'> {
    userID: Scalars['ID'] | null
    headerLine: JSX.Element
    queryUserBatchChangesCodeHosts?: typeof _queryUserBatchChangesCodeHosts
    queryGlobalBatchChangesCodeHosts?: typeof _queryGlobalBatchChangesCodeHosts
}

export const CodeHostConnections: React.FunctionComponent<CodeHostConnectionsProps> = ({
    userID,
    headerLine,
    history,
    location,
    queryUserBatchChangesCodeHosts = _queryUserBatchChangesCodeHosts,
    queryGlobalBatchChangesCodeHosts = _queryGlobalBatchChangesCodeHosts,
}) => {
    // Subject to fire a reload of the list.
    const updateList = useMemo(() => new Subject<void>(), [])
    const query = useCallback<
        (args: Partial<UserBatchChangesCodeHostsVariables>) => Observable<BatchChangesCodeHostsFields>
    >(
        args =>
            userID
                ? queryUserBatchChangesCodeHosts({
                      user: userID,
                      first: args.first ?? null,
                      after: args.after ?? null,
                  })
                : queryGlobalBatchChangesCodeHosts({
                      first: args.first ?? null,
                      after: args.after ?? null,
                  }),
        [queryUserBatchChangesCodeHosts, queryGlobalBatchChangesCodeHosts, userID]
    )
    return (
        <>
            <PageHeader headingElement="h2" path={[{ text: 'Batch Changes' }]} className="mb-3" />
            <Container>
                <h3>Code host tokens</h3>
                {headerLine}
                <FilteredConnection<BatchChangesCodeHostFields, Omit<CodeHostConnectionNodeProps, 'node'>>
                    history={history}
                    location={location}
                    useURLQuery={false}
                    nodeComponent={CodeHostConnectionNode}
                    nodeComponentProps={{ userID, updateList }}
                    queryConnection={query}
                    hideSearch={true}
                    defaultFirst={15}
                    noun="code host"
                    pluralNoun="code hosts"
                    listClassName="list-group"
                    updates={updateList}
                    className="mb-3"
                    cursorPaging={true}
                    noSummaryIfAllNodesVisible={true}
                />
                <p className="mb-0">
                    Code host not present? Site admins can add a code host in{' '}
                    <a
                        href="https://docs.sourcegraph.com/admin/external_service"
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        the manage repositories settings
                    </a>
                    .
                </p>
            </Container>
        </>
    )
}
