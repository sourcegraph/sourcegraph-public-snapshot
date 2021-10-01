import classnames from 'classnames'
import PlusIcon from 'mdi-react/PlusIcon'
import React, { useState } from 'react'
import { useLocation } from 'react-router'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { dataOrThrowErrors } from '@sourcegraph/shared/src/graphql/graphql'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useConnection } from '@sourcegraph/web/src/components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '@sourcegraph/web/src/components/FilteredConnection/ui'
import { Container } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { CodeMonitorFields, ListUserCodeMonitorsResult, ListUserCodeMonitorsVariables } from '../../graphql-operations'
import { Settings } from '../../schema/settings.schema'

import { ListUserCodeMonitors } from './backend'
import { CodeMonitorNode } from './CodeMonitoringNode'
import { CodeMonitoringPageProps } from './CodeMonitoringPage'
import { CodeMonitorSignUpLink } from './CodeMonitoringSignUpLink'

type CodeMonitorFilter = 'all' | 'user'

interface CodeMonitorListProps
    extends Required<Pick<CodeMonitoringPageProps, 'toggleCodeMonitorEnabled'>>,
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

const BATCH_COUNT = 10

export const CodeMonitorList: React.FunctionComponent<CodeMonitorListProps> = ({
    authenticatedUser,
    settingsCascade,
    toggleCodeMonitorEnabled,
}) => {
    const location = useLocation()
    const [monitorListFilter, setMonitorListFilter] = useState<CodeMonitorFilter>('all')

    const { connection: queryConnection, loading, error, fetchMore, hasNextPage } = useConnection<
        ListUserCodeMonitorsResult,
        ListUserCodeMonitorsVariables,
        CodeMonitorFields
    >({
        query: ListUserCodeMonitors,
        variables: {
            // We skip this query when authenticateUser is falsy
            id: authenticatedUser!.id,
            first: BATCH_COUNT,
            after: null,
        },
        options: {
            skip: !authenticatedUser,
            useURL: true,
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)
            if (!data.node) {
                throw new Error('namespace not found')
            }

            if (data.node.__typename !== 'User') {
                throw new Error(`Requested node is a ${data.node.__typename}, not a User or Org`)
            }

            return data.node.monitors
        },
    })

    const connection = authenticatedUser
        ? queryConnection
        : { totalCount: 0, nodes: [], pageInfo: { endCursor: null, hasNextPage: false } }

    const summary = connection && (
        <ConnectionSummary
            connection={connection}
            first={BATCH_COUNT}
            noun="code monitor"
            pluralNoun="code monitors"
            hasNextPage={hasNextPage}
            noSummaryIfAllNodesVisible={true}
            emptyElement={<CodeMonitorEmptyList authenticatedUser={authenticatedUser} />}
            centered={true}
        />
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
                        <ConnectionContainer>
                            {error && <ConnectionError errors={[error.message]} />}
                            {connection && (
                                <ConnectionList>
                                    {connection.nodes.map(node => (
                                        <CodeMonitorNode
                                            key={node.id}
                                            node={node}
                                            isSiteAdminUser={authenticatedUser?.siteAdmin ?? false}
                                            location={location}
                                            toggleCodeMonitorEnabled={toggleCodeMonitorEnabled}
                                            showCodeMonitoringTestEmailButton={
                                                (!isErrorLike(settingsCascade.final) &&
                                                    settingsCascade.final?.experimentalFeatures
                                                        ?.showCodeMonitoringTestEmailButton) ||
                                                false
                                            }
                                        />
                                    ))}
                                </ConnectionList>
                            )}
                            {loading && <ConnectionLoading />}
                            {!loading && connection && (
                                <SummaryContainer centered={true}>
                                    {summary}
                                    {hasNextPage && <ShowMoreButton onClick={fetchMore} centered={true} />}
                                </SummaryContainer>
                            )}
                        </ConnectionContainer>
                    </Container>
                </div>
            </div>
            <div className="mt-5">
                We want to hear your feedback! <a href="mailto:feedback@sourcegraph.com">Share your thoughts</a>
            </div>
        </>
    )
}
