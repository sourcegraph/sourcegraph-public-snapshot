import React, { useEffect } from 'react'

import * as H from 'history'
import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs'
import { startWith, catchError, tap } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { PageHeader, Link, LoadingSpinner, useObservable } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { CodeMonitoringLogo } from '../../code-monitoring/CodeMonitoringLogo'
import { PageTitle } from '../../components/PageTitle'
import { CodeMonitorFields } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import { convertActionsForUpdate } from './action-converters'
import {
    fetchCodeMonitor as _fetchCodeMonitor,
    updateCodeMonitor as _updateCodeMonitor,
    deleteCodeMonitor as _deleteCodeMonitor,
} from './backend'
import { CodeMonitorForm } from './components/CodeMonitorForm'

interface ManageCodeMonitorPageProps extends RouteComponentProps<{ id: Scalars['ID'] }>, ThemeProps {
    authenticatedUser: AuthenticatedUser
    location: H.Location
    history: H.History

    fetchCodeMonitor?: typeof _fetchCodeMonitor
    updateCodeMonitor?: typeof _updateCodeMonitor
    deleteCodeMonitor?: typeof _deleteCodeMonitor

    isSourcegraphDotCom: boolean
}

const AuthenticatedManageCodeMonitorPage: React.FunctionComponent<
    React.PropsWithChildren<ManageCodeMonitorPageProps>
> = ({
    authenticatedUser,
    history,
    location,
    match,
    fetchCodeMonitor = _fetchCodeMonitor,
    updateCodeMonitor = _updateCodeMonitor,
    deleteCodeMonitor = _deleteCodeMonitor,
    isLightTheme,
    isSourcegraphDotCom,
}) => {
    const LOADING = 'loading' as const

    useEffect(() => eventLogger.logViewEvent('ManageCodeMonitorPage'), [])

    const [codeMonitorState, setCodeMonitorState] = React.useState<CodeMonitorFields>({
        id: '',
        description: '',
        enabled: true,
        trigger: { id: '', query: '' },
        actions: { nodes: [] },
    })

    const codeMonitorOrError = useObservable(
        React.useMemo(
            () =>
                fetchCodeMonitor(match.params.id).pipe(
                    tap(monitor => {
                        if (monitor.node !== null && monitor.node.__typename === 'Monitor') {
                            setCodeMonitorState(monitor.node)
                        }
                    }),
                    startWith(LOADING),
                    catchError(error => [asError(error)])
                ),
            [match.params.id, fetchCodeMonitor]
        )
    )

    const updateMonitorRequest = React.useCallback(
        (codeMonitor: CodeMonitorFields): Observable<Partial<CodeMonitorFields>> =>
            updateCodeMonitor(
                {
                    id: match.params.id,
                    update: {
                        namespace: authenticatedUser.id,
                        description: codeMonitor.description,
                        enabled: codeMonitor.enabled,
                    },
                },
                { id: codeMonitor.trigger.id, update: { query: codeMonitor.trigger.query } },
                convertActionsForUpdate(codeMonitor.actions.nodes, authenticatedUser.id)
            ),
        [authenticatedUser.id, match.params.id, updateCodeMonitor]
    )

    return (
        <div className="container col-8">
            <PageTitle title="Manage code monitor" />
            <PageHeader
                path={[{ icon: CodeMonitoringLogo, to: '/code-monitoring' }, { text: 'Manage code monitor' }]}
                description={
                    <>
                        Code monitors watch your code for specific triggers and run actions in response.{' '}
                        <Link to="/help/code_monitoring" target="_blank" rel="noopener">
                            Learn more
                        </Link>
                    </>
                }
            />
            {codeMonitorOrError === 'loading' && <LoadingSpinner />}
            {codeMonitorOrError && !isErrorLike(codeMonitorOrError) && codeMonitorOrError !== 'loading' && (
                <>
                    <CodeMonitorForm
                        history={history}
                        location={location}
                        authenticatedUser={authenticatedUser}
                        deleteCodeMonitor={deleteCodeMonitor}
                        onSubmit={updateMonitorRequest}
                        codeMonitor={codeMonitorState}
                        submitButtonLabel="Save"
                        showDeleteButton={true}
                        isLightTheme={isLightTheme}
                        isSourcegraphDotCom={isSourcegraphDotCom}
                    />
                </>
            )}
        </div>
    )
}

export const ManageCodeMonitorPage = withAuthenticatedUser(AuthenticatedManageCodeMonitorPage)
