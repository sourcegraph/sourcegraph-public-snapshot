import PlusIcon from 'mdi-react/PlusIcon'
import React, { useCallback, useState } from 'react'
import { useHistory, useLocation } from 'react-router'
import { of } from 'rxjs'

import { isErrorLike } from '@sourcegraph/common'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { Container, Button } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { FilteredConnection } from '../../components/FilteredConnection'
import { CodeMonitorFields, ListUserCodeMonitorsResult, ListUserCodeMonitorsVariables } from '../../graphql-operations'
import { Settings } from '../../schema/settings.schema'

import { CodeMonitorInfo } from './CodeMonitorInfo'
import { CodeMonitorNode, CodeMonitorNodeProps } from './CodeMonitoringNode'
import { CodeMonitoringPageProps } from './CodeMonitoringPage'
import { CodeMonitorSignUpLink } from './CodeMonitoringSignUpLink'

type CodeMonitorFilter = 'all' | 'user'

interface CodeMonitorListProps
    extends Required<Pick<CodeMonitoringPageProps, 'fetchUserCodeMonitors' | 'toggleCodeMonitorEnabled'>>,
        SettingsCascadeProps<Settings> {
    authenticatedUser: AuthenticatedUser | null
}

const CodeMonitorEmptyList: React.FunctionComponent<{ authenticatedUser: AuthenticatedUser | null }> = ({
    authenticatedUser,
}) => (
    <div className="text-center">
        <h2 className="text-muted mb-2">No code monitors have been created.</h2>
        {authenticatedUser ? (
            <Link to="/code-monitoring/new" className="btn btn-primary">
                <PlusIcon className="icon-inline" />
                Create a code monitor
            </Link>
        ) : (
            <CodeMonitorSignUpLink eventName="SignUpPLGMonitor_EmptyList" text="Sign up to create a code monitor" />
        )}
    </div>
)

export const CodeMonitorList: React.FunctionComponent<CodeMonitorListProps> = ({
    authenticatedUser,
    settingsCascade,
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
                    <h3>Filters</h3>
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
                    <h3 className="mb-2">
                        {`${monitorListFilter === 'all' ? 'All code monitors' : 'Your code monitors'}`}
                    </h3>
                    <Container>
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
                                isSiteAdminUser: authenticatedUser?.siteAdmin ?? false,
                                location,
                                showCodeMonitoringTestEmailButton:
                                    (!isErrorLike(settingsCascade.final) &&
                                        settingsCascade.final?.experimentalFeatures
                                            ?.showCodeMonitoringTestEmailButton) ||
                                    false,
                                toggleCodeMonitorEnabled,
                            }}
                            noun="code monitor"
                            pluralNoun="code monitors"
                            noSummaryIfAllNodesVisible={true}
                            cursorPaging={true}
                            withCenteredSummary={true}
                            emptyElement={<CodeMonitorEmptyList authenticatedUser={authenticatedUser} />}
                        />
                    </Container>
                </div>
            </div>
            <div className="mt-5">
                We want to hear your feedback! <a href="mailto:feedback@sourcegraph.com">Share your thoughts</a>
            </div>
        </>
    )
}
