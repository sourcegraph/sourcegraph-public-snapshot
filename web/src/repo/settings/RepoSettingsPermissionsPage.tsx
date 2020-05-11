import React, { useEffect } from 'react'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PageTitle } from '../../components/PageTitle'
import { Timestamp } from '../../components/time/Timestamp'
import { eventLogger } from '../../tracking/eventLogger'

/**
 * The repository settings permissions page.
 */
export const RepoSettingsPermissionsPage: React.FunctionComponent<{ repo: GQL.IRepository }> = ({ repo }) => {
    useEffect(() => eventLogger.logViewEvent('RepoSettingsPermissions'))

    return (
        <div className="repo-settings-permissions-page w-100">
            <PageTitle title="Permissions" />
            <h2>Permissions</h2>
            {!repo.isPrivate ? (
                <div className="alert alert-info">
                    Access to this repository is not restricted, all Sourcegraph users have access.
                </div>
            ) : !repo.permissionsInfo ? (
                <div className="alert alert-info">
                    This repository is queued to sync permissions, only site admins will have access to it until syncing
                    is finished.
                </div>
            ) : (
                <table className="table">
                    <tbody>
                        <tr>
                            <th>Last complete sync</th>
                            <td>
                                {repo.permissionsInfo.syncedAt ? (
                                    <Timestamp date={repo.permissionsInfo.syncedAt} />
                                ) : (
                                    'Never'
                                )}
                            </td>
                            <td className="text-muted">Updated by repository permissions syncing</td>
                        </tr>
                        <tr>
                            <th>Last incremental sync</th>
                            <td>
                                <Timestamp date={repo.permissionsInfo.updatedAt} />
                            </td>
                            <td className="text-muted">Updated by user permissions syncing</td>
                        </tr>
                    </tbody>
                </table>
            )}
        </div>
    )
}
