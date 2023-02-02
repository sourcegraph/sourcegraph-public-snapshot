import React, { useCallback } from 'react'

import { useLocation } from 'react-router'
import { of } from 'rxjs'

import { buildCloudTrialURL } from '@sourcegraph/shared/src/util/url'
import { Container, Link, H2, H3 } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { CloudCtaBanner } from '../../components/CloudCtaBanner'
import { FilteredConnection } from '../../components/FilteredConnection'
import { CodeMonitorFields, ListUserCodeMonitorsResult, ListUserCodeMonitorsVariables } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import { CodeMonitorNode, CodeMonitorNodeProps } from './CodeMonitoringNode'
import { CodeMonitoringPageProps } from './CodeMonitoringPage'

interface CodeMonitorListProps
    extends Required<Pick<CodeMonitoringPageProps, 'fetchUserCodeMonitors' | 'toggleCodeMonitorEnabled'>> {
    authenticatedUser: AuthenticatedUser | null
}

const CodeMonitorEmptyList: React.FunctionComponent<
    React.PropsWithChildren<{ authenticatedUser: AuthenticatedUser | null }>
> = ({ authenticatedUser }) => (
    <div className="text-center">
        <H2 className="text-muted mb-2">No code monitors have been created.</H2>
    </div>
)

export const CodeMonitorList: React.FunctionComponent<React.PropsWithChildren<CodeMonitorListProps>> = ({
    authenticatedUser,
    fetchUserCodeMonitors,
    toggleCodeMonitorEnabled,
}) => {
    const location = useLocation()
    const isSourcegraphDotCom: boolean = window.context?.sourcegraphDotComMode || false

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

    return (
        <>
            <div className="row mb-5">
                <div className="d-flex flex-column w-100 col">
                    <div className="d-flex align-items-center justify-content-between">
                        <H3 className="mb-2">Your code monitors</H3>
                        {isSourcegraphDotCom && (
                            <CloudCtaBanner variant="outlined" small={true}>
                                To monitor changes across your private repos,{' '}
                                <Link
                                    to={buildCloudTrialURL(authenticatedUser, 'monitoring')}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    onClick={() => eventLogger.log('ClickedOnCloudCTA', { cloudCtaType: 'Monitoring' })}
                                >
                                    try Sourcegraph Cloud
                                </Link>
                                .
                            </CloudCtaBanner>
                        )}
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
                            }}
                            noun="code monitor"
                            pluralNoun="code monitors"
                            noSummaryIfAllNodesVisible={true}
                            cursorPaging={true}
                            withCenteredSummary={true}
                            emptyElement={<CodeMonitorEmptyList authenticatedUser={authenticatedUser} />}
                            listComponent="div"
                        />
                    </Container>
                </div>
            </div>
            <div className="mt-5">
                We want to hear your feedback! <Link to="mailto:feedback@sourcegraph.com">Share your thoughts</Link>
            </div>
        </>
    )
}
