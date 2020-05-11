import React, { useEffect } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PageTitle } from '../../../components/PageTitle'
import { Timestamp } from '../../../components/time/Timestamp'
import { eventLogger } from '../../../tracking/eventLogger'

/**
 * The user settings permissions page.
 */
export const UserSettingsPermissionsPage: React.FunctionComponent<{ user: GQL.IUser }> = ({ user }) => {
    useEffect(() => eventLogger.logViewEvent('UserSettingsPermissions'))

    return (
        <div className="w-100">
            <PageTitle title="Permissions" />
            <h2>Permissions</h2>
            {user.siteAdmin ? (
                <div className="alert alert-info">
                    Site admin can access all repositories in the Sourcegraph instance.
                </div>
            ) : !user.permissionsInfo ? (
                <div className="alert alert-info">
                    This user is queued to sync permissions, it can only access non-private repositories until syncing
                    is finished.
                </div>
            ) : (
                <table className="table">
                    <tbody>
                        <tr>
                            <th>Last complete sync</th>
                            <td>
                                {user.permissionsInfo.syncedAt ? (
                                    <Timestamp date={user.permissionsInfo.syncedAt} />
                                ) : (
                                    'Never'
                                )}
                            </td>
                            <td className="text-muted">Updated by user permissions syncing</td>
                        </tr>
                        <tr>
                            <th>Last incremental sync</th>
                            <td>
                                <Timestamp date={user.permissionsInfo.updatedAt} />
                            </td>
                            <td className="text-muted">Updated by repository permissions syncing</td>
                        </tr>
                    </tbody>
                </table>
            )}
        </div>
    )
}
