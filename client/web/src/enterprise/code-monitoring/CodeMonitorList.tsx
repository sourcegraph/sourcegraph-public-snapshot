import React, { useCallback, useState } from 'react'

import { useHistory, useLocation } from 'react-router'
import { of } from 'rxjs'

import { Button, Container, Link, Typography } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { FilteredConnection } from '../../components/FilteredConnection'
import { CodeMonitorFields, ListUserCodeMonitorsResult, ListUserCodeMonitorsVariables } from '../../graphql-operations'

import { CodeMonitorInfo } from './CodeMonitorInfo'
import { CodeMonitorNode, CodeMonitorNodeProps } from './CodeMonitoringNode'
import { CodeMonitoringPageProps } from './CodeMonitoringPage'
import { CodeMonitorSignUpLink } from './CodeMonitoringSignUpLink'

type CodeMonitorFilter = 'all' | 'user'

interface CodeMonitorListProps
    extends Required<Pick<CodeMonitoringPageProps, 'fetchUserCodeMonitors' | 'toggleCodeMonitorEnabled'>> {
    authenticatedUser: AuthenticatedUser | null
}

const CodeMonitorEmptyList: React.FunctionComponent<
    React.PropsWithChildren<{ authenticatedUser: AuthenticatedUser | null }>
> = ({ authenticatedUser }) => (
    <div className="text-center">
        <Typography.H2 className="text-muted mb-2">No code monitors have been created.</Typography.H2>
        {!authenticatedUser && (
            <CodeMonitorSignUpLink
                className="my-3"
                eventName="SignUpPLGMonitor_EmptyList"
                text="Get started with code monitors"
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
    const [monitorListFilter, setMonitorListFilter] = useState<CodeMonitorFilter>('all')

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
                <div className="d-flex flex-column col-2 mr-2">
                    <Typography.H3 as={Typography.H2}>Filters</Typography.H3>
                    <Button
                        className="text-left"
                        onClick={() => setMonitorListFilter('all')}
                        variant={monitorListFilter === 'all' ? 'primary' : undefined}
                    >
                        All
                    </Button>
                    <Button
                        className="text-left"
                        onClick={() => setMonitorListFilter('user')}
                        variant={monitorListFilter === 'user' ? 'primary' : undefined}
                    >
                        Your code monitors
                    </Button>
                </div>
                <div className="d-flex flex-column w-100 col">
                    <CodeMonitorInfo />
                    <Typography.H3 className="mb-2">
                        {`${monitorListFilter === 'all' ? 'All code monitors' : 'Your code monitors'}`}
                    </Typography.H3>
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
