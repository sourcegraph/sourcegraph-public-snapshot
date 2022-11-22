import React, { useCallback } from 'react'

import { useHistory, useLocation } from 'react-router'
import { of } from 'rxjs'

import { Container, Link, H2, H3 } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { FilteredConnection } from '../../components/FilteredConnection'
import { CodeMonitorFields, ListUserCodeMonitorsResult, ListUserCodeMonitorsVariables } from '../../graphql-operations'

import { CodeMonitorNode, CodeMonitorNodeProps } from './CodeMonitoringNode'
import { CodeMonitoringPageProps } from './CodeMonitoringPage'
import { CodeMonitorSignUpLink } from './CodeMonitoringSignUpLink'

interface CodeMonitorListProps
    extends Required<Pick<CodeMonitoringPageProps, 'fetchUserCodeMonitors' | 'toggleCodeMonitorEnabled'>> {
    authenticatedUser: AuthenticatedUser | null
}

const CodeMonitorEmptyList: React.FunctionComponent<
    React.PropsWithChildren<{ authenticatedUser: AuthenticatedUser | null }>
> = ({ authenticatedUser }) => (
    <div className="text-center">
        <H2 className="text-muted mb-2">No code monitors have been created.</H2>
        {!authenticatedUser && (
            <CodeMonitorSignUpLink
                className="my-3"
                eventName="SignUpPLGMonitor_EmptyList"
                text="Get started with code monitors"
                cloudSignup={true}
            />
        )}
    </div>
)

export const CodeMonitorList: React.FunctionComponent<React.PropsWithChildren<CodeMonitorListProps>> = ({
    authenticatedUser,
    fetchUserCodeMonitors,
    toggleCodeMonitorEnabled,
}) => {
    const location = useLocation()
    const history = useHistory()

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
                    <H3 className="mb-2">Your code monitors</H3>
                    <Container className="py-3">
                        <FilteredConnection<
                            CodeMonitorFields,
                            Omit<CodeMonitorNodeProps, 'node'>,
                            (ListUserCodeMonitorsResult['node'] & { __typename: 'User' })['monitors']
                        >
                            location={location}
                            history={history}
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
