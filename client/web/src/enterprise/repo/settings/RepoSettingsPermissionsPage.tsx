import React, { useEffect, useMemo } from 'react'

import * as H from 'history'

import { Container, PageHeader, LoadingSpinner, useObservable, Alert, Link } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../components/PageTitle'
import { Timestamp } from '../../../components/time/Timestamp'
import { SettingsAreaRepositoryFields } from '../../../graphql-operations'
import { ActionContainer } from '../../../repo/settings/components/ActionContainer'
import { scheduleRepositoryPermissionsSync } from '../../../site-admin/backend'
import { eventLogger } from '../../../tracking/eventLogger'

import { repoPermissionsInfo } from './backend'

export interface RepoSettingsPermissionsPageProps {
    repo: SettingsAreaRepositoryFields
    history: H.History
}

/**
 * The repository settings permissions page.
 */
export const RepoSettingsPermissionsPage: React.FunctionComponent<
    React.PropsWithChildren<RepoSettingsPermissionsPageProps>
> = ({ repo, history }) => {
    useEffect(() => eventLogger.logViewEvent('RepoSettingsPermissions'))
    const permissionsInfo = useObservable(useMemo(() => repoPermissionsInfo(repo.id), [repo.id]))

    if (permissionsInfo === undefined) {
        return <LoadingSpinner inline={false} />
    }

    return (
        <>
            <PageTitle title="Permissions" />
            <PageHeader
                path={[{ text: 'Permissions' }]}
                headingElement="h2"
                className="mb-3"
                description={
                    <>
                        Learn more about{' '}
                        <Link to="/help/admin/repo/permissions#background-permissions-syncing">
                            background permissions syncing
                        </Link>
                        .
                    </>
                }
            />
            <Container className="repo-settings-permissions-page">
                {!repo.isPrivate ? (
                    <Alert className="mb-0" variant="info">
                        Access to this repository is <strong>not restricted</strong>, all Sourcegraph users have access.
                    </Alert>
                ) : !permissionsInfo ? (
                    <Alert className="mb-0" variant="info">
                        This repository is queued to sync permissions, only site admins will have access to it until
                        syncing is finished.
                    </Alert>
                ) : permissionsInfo.unrestricted ? (
                    <Alert className="mb-0" variant="info">
                        This repository has been explicitly flagged as unrestricted, all Sourcegraph users have access.
                    </Alert>
                ) : (
                    <div>
                        <table className="table">
                            <tbody>
                                <tr>
                                    <th>Last complete sync</th>
                                    <td>
                                        {permissionsInfo.syncedAt ? (
                                            <Timestamp date={permissionsInfo.syncedAt} />
                                        ) : (
                                            'Never'
                                        )}
                                    </td>
                                    <td className="text-muted">Updated by repository permissions syncing</td>
                                </tr>
                                <tr>
                                    <th>Last incremental sync</th>
                                    <td>
                                        <Timestamp date={permissionsInfo.updatedAt} />
                                    </td>
                                    <td className="text-muted">Updated by user permissions syncing</td>
                                </tr>
                            </tbody>
                        </table>
                        <ScheduleRepositoryPermissionsSyncActionContainer repo={repo} history={history} />
                    </div>
                )}
            </Container>
        </>
    )
}

interface ScheduleRepositoryPermissionsSyncActionContainerProps {
    repo: SettingsAreaRepositoryFields
    history: H.History
}

class ScheduleRepositoryPermissionsSyncActionContainer extends React.PureComponent<ScheduleRepositoryPermissionsSyncActionContainerProps> {
    public render(): JSX.Element | null {
        return (
            <ActionContainer
                title="Manually schedule a permissions sync"
                description={
                    <div>Site admins are able to manually schedule a permissions sync for this repository.</div>
                }
                buttonLabel="Schedule now"
                flashText="Added to queue"
                run={this.scheduleRepositoryPermissions}
                history={this.props.history}
            />
        )
    }

    private scheduleRepositoryPermissions = async (): Promise<void> => {
        await scheduleRepositoryPermissionsSync({ repository: this.props.repo.id }).toPromise()
    }
}
