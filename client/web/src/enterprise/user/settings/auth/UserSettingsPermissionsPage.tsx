import React, { useEffect } from 'react'

import { mdiInformationOutline } from '@mdi/js'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { RepoLink } from '@sourcegraph/shared/src/components/RepoLink'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Container,
    useDebounce,
    PageSwitcher,
    PageHeader,
    LoadingSpinner,
    Input,
    Link,
    Badge,
    BadgeProps,
    Tooltip,
    Icon,
} from '@sourcegraph/wildcard'

import { usePageSwitcherPagination } from '../../../../components/FilteredConnection/hooks/usePageSwitcherPagination'
import { PageTitle } from '../../../../components/PageTitle'
import {
    UserPermissionsInfoResult,
    UserPermissionsInfoVariables,
    PermissionsInfoRepositoryFields as INode,
    UserPermissionsInfoUserNode as IUser,
} from '../../../../graphql-operations'
import { useURLSyncedState } from '../../../../hooks'
import { ActionContainer } from '../../../../repo/settings/components/ActionContainer'
import { ExternalRepositoryIcon } from '../../../../site-admin/components/ExternalRepositoryIcon'
import { PermissionsSyncJobsTable } from '../../../../site-admin/permissions-center/PermissionsSyncJobsTable'
import { Table, IColumn } from '../../../../site-admin/UserManagement/components/Table'
import { eventLogger } from '../../../../tracking/eventLogger'

import { scheduleUserPermissionsSync, UserPermissionsInfoQuery } from './backend'

import styles from './UserSettingsPermissionsPage.module.scss'

interface Props extends TelemetryProps {
    user: { id: string; username: string }
}

/**
 * The user settings permissions page.
 */
export const UserSettingsPermissionsPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    user,
    telemetryService,
}) => {
    useEffect(() => eventLogger.logViewEvent('UserSettingsPermissions'), [])

    const [{ query }, setSearchQuery] = useURLSyncedState({ query: '' })
    const debouncedQuery = useDebounce(query, 300)

    const { connection, data, loading, error, refetch, variables, ...paginationProps } = usePageSwitcherPagination<
        UserPermissionsInfoResult,
        UserPermissionsInfoVariables,
        INode
    >({
        query: UserPermissionsInfoQuery,
        variables: {
            userID: user.id,
            query: debouncedQuery,
        },
        getConnection: ({ data }) => {
            if (data?.node?.__typename === 'User') {
                const node = data.node as IUser

                return node.permissionsInfo?.repositories || undefined
            }

            return undefined
        },
    })

    useEffect(() => {
        if (debouncedQuery !== variables.query) {
            refetch({ ...variables, query: debouncedQuery })
        }
    }, [debouncedQuery, refetch, variables])

    const permissionsInfo = data?.node?.__typename === 'User' ? (data.node as IUser).permissionsInfo : undefined

    if (!permissionsInfo || !connection) {
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
                <>
                    <table className="table">
                        <tbody>
                            <tr>
                                <th className="border-0">
                                    Last complete sync{' '}
                                    <Tooltip content="Syncs user permissions from the code host. All repositories that the user has access to on the code host will be accessible on Sourcegraph as well.">
                                        <Icon aria-label="more-info" svgPath={mdiInformationOutline} />
                                    </Tooltip>
                                </th>
                                <td className="border-0">
                                    {permissionsInfo.syncedAt ? <Timestamp date={permissionsInfo.syncedAt} /> : 'Never'}
                                </td>
                                <td className="text-muted border-0">Updated by user-centric permission sync.</td>
                            </tr>
                            <tr>
                                <th>
                                    Last partial sync{' '}
                                    <Tooltip content="Syncs repository permissions from the code host. If a repository-centric sync returns this user as accessor, it is noted here as partial sync. Partial syncs do not show in the list of permission sync jobs below.">
                                        <Icon aria-label="more-info" svgPath={mdiInformationOutline} />
                                    </Tooltip>
                                </th>
                                <td>
                                    {permissionsInfo.updatedAt === null ? (
                                        'Never'
                                    ) : (
                                        <Timestamp date={permissionsInfo.updatedAt} />
                                    )}
                                </td>
                                <td className="text-muted">
                                    Partial update done by repositor-centric permission sync.
                                </td>
                            </tr>
                        </tbody>
                    </table>
                    <ScheduleUserPermissionsSyncActionContainer user={user} />
                </>
            </Container>
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Permissions Sync Jobs' }]}
                description={
                    <>
                        List of permissions sync jobs. A permission sync job fetches the newest permissions for the
                        given user.
                    </>
                }
                className="my-3 pt-3"
            />
            <Container className="mb-3">
                <PermissionsSyncJobsTable telemetryService={telemetryService} minimal={true} userID={user.id} />
            </Container>
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Accessible Repositories' }]}
                description="List of repositories which are accessible to the user."
                className="my-3 pt-3"
            />
            <Container className="mb-3">
                <div className="d-flex mb-3">
                    <Input
                        type="search"
                        placeholder="Search repositories..."
                        name="query"
                        value={query}
                        onChange={event => setSearchQuery({ query: event.currentTarget.value })}
                        autoComplete="off"
                        autoCorrect="off"
                        autoCapitalize="off"
                        spellCheck={false}
                        aria-label="Search repositories..."
                        variant="regular"
                    />
                    <div className="flex-1" />
                </div>
                <Table<INode> columns={TableColumns} getRowId={node => node.id} data={connection.nodes} />
                <PageSwitcher
                    {...paginationProps}
                    className="mt-4"
                    totalCount={connection?.totalCount ?? null}
                    totalLabel="accessible repositories"
                />
            </Container>
        </div>
    )
}

const TableColumns: IColumn<INode>[] = [
    {
        key: 'repository',
        header: 'Repository',
        render: ({ repository }: INode) =>
            repository && (
                <div key={repository.id} className="py-2">
                    <ExternalRepositoryIcon externalRepo={repository.externalRepository} />
                    <RepoLink repoName={repository.name} to={repository.url} />
                </div>
            ),
    },
    {
        key: 'reason',
        header: 'Reason',
        render: ({ reason }: INode) => <Badge {...(PermissionReasonBadgeProps[reason] || {})}>{reason}</Badge>,
    },
    {
        key: 'updatedAt',
        header: 'Updated At',
        render: ({ updatedAt }: INode) => (
            <div className={styles.updatedAtCell}>{updatedAt ? <Timestamp date={updatedAt} /> : '-'}</div>
        ),
    },
]

const PermissionReasonBadgeProps: { [reason: string]: BadgeProps } = {
    'Permissions Sync': {
        variant: 'success',
        tooltip: 'The repository is accessible to the user due to permissions syncing from code host.',
    },
    Unrestricted: { variant: 'primary', tooltip: 'The repository is accessible to all the users. ' },
    'Site Admin': { variant: 'secondary', tooltip: 'The user is site admin and has access to all the repositories.' },
}

interface ScheduleUserPermissionsSyncActionContainerProps {
    user: { id: string; username: string }
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
