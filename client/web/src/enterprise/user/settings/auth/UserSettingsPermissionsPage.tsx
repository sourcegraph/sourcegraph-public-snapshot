import React, { useEffect, useMemo } from 'react'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { Container, PageHeader, LoadingSpinner, useObservable, Alert, Link } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../../components/PageTitle'
import { UserSettingsAreaUserFields } from '../../../../graphql-operations'
import { ActionContainer } from '../../../../repo/settings/components/ActionContainer'
import { eventLogger } from '../../../../tracking/eventLogger'

import { scheduleUserPermissionsSync, userPermissionsInfo } from './backend'

import styles from './UserSettingsPermissionsPage.module.scss'

/**
 * The user settings permissions page.
 */
export const UserSettingsPermissionsPage: React.FunctionComponent<
    React.PropsWithChildren<{
        user: UserSettingsAreaUserFields
    }>
> = ({ user }) => {
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
                        Learn more about <Link to="/help/admin/permissions/syncing">permissions syncing</Link>.
                    </>
                }
                className="mb-3"
            />
            <Container className="mb-3">
                {user.siteAdmin && !window.context.site['authz.enforceForSiteAdmins'] ? (
                    <Alert className="mb-0" variant="info">
                        Site admin can access all repositories in the Sourcegraph instance.
                    </Alert>
                ) : !permissionsInfo ? (
                    <Alert className="mb-0" variant="info">
                        This user is queued to sync permissions, it can only access non-private repositories until
                        syncing is finished.
                    </Alert>
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
                        <ScheduleUserPermissionsSyncActionContainer user={user} />
                    </>
                )}
            </Container>
        </div>
    )
}

interface ScheduleUserPermissionsSyncActionContainerProps {
    user: UserSettingsAreaUserFields
}

class ScheduleUserPermissionsSyncActionContainer extends React.PureComponent<ScheduleUserPermissionsSyncActionContainerProps> {
    public render(): JSX.Element | null {
        return (
            <ActionContainer
                title="Manually schedule a permissions sync"
                description={<div>Schedule a permissions sync for user: {this.props.user.username}.</div>}
                buttonLabel="Schedule now"
                flashText="Added to queue"
                run={this.scheduleUserPermissions}
                className={styles.userSettingsPermissionsPageSyncContainer}
            />
        )
    }

    private scheduleUserPermissions = async (): Promise<void> => {
        await scheduleUserPermissionsSync({ user: this.props.user.id }).toPromise()
    }
}
