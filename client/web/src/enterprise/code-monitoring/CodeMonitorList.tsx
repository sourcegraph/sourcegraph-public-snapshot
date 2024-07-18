import React, { useCallback } from 'react'

import { useLocation } from 'react-router-dom'
import { of } from 'rxjs'

import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Container, H2, H3 } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { FilteredConnection } from '../../components/FilteredConnection'
import type {
    CodeMonitorFields,
    ListAllCodeMonitorsResult,
    ListAllCodeMonitorsVariables,
    ListUserCodeMonitorsResult,
    ListUserCodeMonitorsVariables,
} from '../../graphql-operations'

import { CodeMonitorNode, type CodeMonitorNodeProps } from './CodeMonitoringNode'
import type { CodeMonitoringPageProps } from './CodeMonitoringPage'

interface CodeMonitorListProps
    extends TelemetryV2Props,
        Required<
            Pick<CodeMonitoringPageProps, 'fetchUserCodeMonitors' | 'fetchCodeMonitors' | 'toggleCodeMonitorEnabled'>
        > {
    authenticatedUser: AuthenticatedUser | null
}

const CodeMonitorEmptyList: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <div className="text-center">
        <H2 className="text-muted mb-2">No code monitors have been created.</H2>
    </div>
)

export const CodeMonitorList: React.FunctionComponent<React.PropsWithChildren<CodeMonitorListProps>> = ({
    authenticatedUser,
    fetchUserCodeMonitors,
    fetchCodeMonitors,
    toggleCodeMonitorEnabled,
}) => {
    const location = useLocation()

    const queryConnection = useCallback(
        (args: Partial<ListUserCodeMonitorsVariables>) => {
            if (!authenticatedUser) {
                return of({
                    totalCount: 0,
                    nodes: [],
                    pageInfo: { endCursor: null, hasNextPage: false },
                })
            }

            return fetchUserCodeMonitors({
                id: authenticatedUser.id,
                first: args.first ?? null,
                after: args.after ?? null,
            })
        },
        [authenticatedUser, fetchUserCodeMonitors]
    )

    const queryAllConnection = useCallback(
        (args: Omit<Partial<ListAllCodeMonitorsVariables>, 'first'> & { first?: number | null }) =>
            fetchCodeMonitors({
                first: args.first ?? 10,
                after: args.after ?? null,
            }),
        [fetchCodeMonitors]
    )

    return (
        <>
            <div className="row mb-5">
                <div className="d-flex flex-column w-100 col">
                    <div className="d-flex align-items-center justify-content-between">
                        <H3 className="mb-2">Your code monitors</H3>
                    </div>
                    <Container className="py-3">
                        <FilteredConnection<
                            CodeMonitorFields,
                            Omit<CodeMonitorNodeProps, 'node'>,
                            (ListUserCodeMonitorsResult['node'] & { __typename: 'User' })['monitors']
                        >
                            defaultFirst={10}
                            queryConnection={queryConnection}
                            hideSearch={true}
                            nodeComponent={CodeMonitorNode}
                            nodeComponentProps={{
                                location,
                                toggleCodeMonitorEnabled,
                                showOwner: false,
                            }}
                            noun="code monitor"
                            pluralNoun="code monitors"
                            noSummaryIfAllNodesVisible={true}
                            cursorPaging={true}
                            withCenteredSummary={true}
                            emptyElement={<CodeMonitorEmptyList />}
                            listComponent="div"
                        />
                    </Container>
                </div>
            </div>
            <div className="row mb-5">
                <div className="d-flex flex-column w-100 col">
                    {authenticatedUser?.siteAdmin && (
                        <>
                            <div className="d-flex align-items-center justify-content-between">
                                <H3 className="mb-2">All code monitors</H3>
                            </div>
                            <Container className="py-3">
                                <FilteredConnection<
                                    CodeMonitorFields,
                                    Omit<CodeMonitorNodeProps, 'node'>,
                                    ListAllCodeMonitorsResult['monitors']['nodes']
                                >
                                    defaultFirst={10}
                                    queryConnection={queryAllConnection}
                                    hideSearch={true}
                                    nodeComponent={CodeMonitorNode}
                                    nodeComponentProps={{
                                        location,
                                        toggleCodeMonitorEnabled,
                                        showOwner: authenticatedUser?.siteAdmin ?? false,
                                    }}
                                    noun="code monitor"
                                    pluralNoun="code monitors"
                                    noSummaryIfAllNodesVisible={true}
                                    cursorPaging={true}
                                    withCenteredSummary={true}
                                    emptyElement={<CodeMonitorEmptyList />}
                                    listComponent="div"
                                />
                            </Container>
                        </>
                    )}
                </div>
            </div>
        </>
    )
}
