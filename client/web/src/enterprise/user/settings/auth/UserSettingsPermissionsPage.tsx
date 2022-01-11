import * as H from 'history'
import React, { useEffect, useMemo } from 'react'

import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Container, PageHeader, LoadingSpinner } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../../components/PageTitle'
import { Timestamp } from '../../../../components/time/Timestamp'
import { UserSettingsAreaUserFields } from '../../../../graphql-operations'
import { ActionContainer } from '../../../../repo/settings/components/ActionContainer'
import { eventLogger } from '../../../../tracking/eventLogger'

import { scheduleUserPermissionsSync, userPermissionsInfo } from './backend'
import styles from './UserSettingsPermissionsPage.module.scss'

/**
 * The user settings permissions page.
 */
export const UserSettingsPermissionsPage: React.FunctionComponent<{
    user: UserSettingsAreaUserFields
    history: H.History
}> = ({ user, history }) => {
    useEffect(() => eventLogger.logViewEvent('UserSettingsPermissions'))
    const permissionsInfo = useObservable(useMemo(() => userPermissionsInfo(user.id), [user.id]))

    if (permissionsInfo === undefined) {
        return <LoadingSpinner />
    }

    return (
        <div className="w-100">
            <PageTitle title="Permissions" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Permissions' }]}
                description={
                    <>
                        Learn more about{' '}
                        <a href="/help/admin/repo/permissions#background-permissions-syncing">
                            background permissions syncing
                        </a>
                        .
                    </>
                }
                className="mb-3"
            />
            <Container className="mb-3">
                {user.siteAdmin && !window.context.site['authz.enforceForSiteAdmins'] ? (
                    <div className="alert alert-info mb-0">
                        Site admin can access all repositories in the Sourcegraph instance.
                    </div>
                ) : !permissionsInfo ? (
                    <div className="alert alert-info mb-0">
                        This user is queued to sync permissions, it can only access non-private repositories until
                        syncing is finished.
                    </div>
                ) : (
                    <>
                        <table className="table">
                            <tbody>
                                <tr>
                                    <th className="border-0">Last complete sync</th>
                                    <td className="border-0">
                                        {permissionsInfo.syncedAt ? (
                                            <Timestamp date={permissionsInfo.syncedAt} />
                                        ) : (
                                            'Never'
                                        )}
                                    </td>
                                    <td className="text-muted border-0">Updated by user permissions syncing</td>
                                </tr>
                                <tr>
                                    <th>Last incremental sync</th>
                                    <td>
                                        <Timestamp date={permissionsInfo.updatedAt} />
                                    </td>
                                    <td className="text-muted">Updated by repository permissions syncing</td>
                                </tr>
                            </tbody>
                        </table>
                        <ScheduleUserPermissionsSyncActionContainer user={user} history={history} />
                    </>
                )}
            </Container>
        </div>
    )
}

interface ScheduleUserPermissionsSyncActionContainerProps {
    user: UserSettingsAreaUserFields
    history: H.History
}

class ScheduleUserPermissionsSyncActionContainer extends React.PureComponent<ScheduleUserPermissionsSyncActionContainerProps> {
    public render(): JSX.Element | null {
        return (
            <ActionContainer
                title="Manually schedule a permissions sync"
                description={<div>Site admins are able to manually schedule a permissions sync for this user.</div>}
                buttonLabel="Schedule now"
                flashText="Added to queue"
                run={this.scheduleUserPermissions}
                history={this.props.history}
                className={styles.userSettingsPermissionsPageSyncContainer}
            />
        )
    }

    private scheduleUserPermissions = async (): Promise<void> => {
        await scheduleUserPermissionsSync({ user: this.props.user.id }).toPromise()
    }
}
