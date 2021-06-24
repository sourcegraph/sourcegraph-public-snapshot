import classnames from 'classnames'
import React, { useCallback, useState } from 'react'
import { useHistory, useLocation } from 'react-router'

import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { Container } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { FilteredConnection } from '../../components/FilteredConnection'
import { CodeMonitorFields, ListUserCodeMonitorsResult, ListUserCodeMonitorsVariables } from '../../graphql-operations'
import { Settings } from '../../schema/settings.schema'

import { CodeMonitorNode, CodeMonitorNodeProps } from './CodeMonitoringNode'
import { CodeMonitoringPageProps } from './CodeMonitoringPage'

type CodeMonitorFilter = 'all' | 'user'

interface CodeMonitorListProps
    extends Required<Pick<CodeMonitoringPageProps, 'fetchUserCodeMonitors' | 'toggleCodeMonitorEnabled'>>,
        SettingsCascadeProps<Settings> {
    authenticatedUser: AuthenticatedUser
}

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
        (args: Partial<ListUserCodeMonitorsVariables>) =>
            fetchUserCodeMonitors({
                id: authenticatedUser.id,
                first: args.first ?? null,
                after: args.after ?? null,
            }),
        [authenticatedUser, fetchUserCodeMonitors]
    )

    return (
        <>
            <div className="row mb-5">
                <div className="d-flex flex-column col-2 mr-2">
                    <h3>Filters</h3>
                    <button
                        type="button"
                        className={classnames('btn text-left', {
                            'btn-primary': monitorListFilter === 'all',
                        })}
                        onClick={() => setMonitorListFilter('all')}
                    >
                        All
                    </button>
                    <button
                        type="button"
                        className={classnames('btn text-left', {
                            'btn-primary': monitorListFilter === 'user',
                        })}
                        onClick={() => setMonitorListFilter('user')}
                    >
                        Your code monitors
                    </button>
                </div>
                <div className="d-flex flex-column w-100 col">
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
                                authenticatedUser,
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
                            className="filtered-connection__centered-summary"
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
