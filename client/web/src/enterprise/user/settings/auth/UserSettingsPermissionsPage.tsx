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
    Icon,
    Text,
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
                    <div className="d-flex">
                        <b>Last update to permissions </b>
                        <span className="d-flex flex-grow-1">
                            {permissionsInfo.updatedAt ? (
                                <>
                                    <span className="flex-grow-1 text-center">
                                        <Timestamp date={permissionsInfo.updatedAt} />
                                    </span>
                                    <span>
                                        by{' '}
                                        <PermissionSource
                                            source={permissionsInfo.source as PermissionSourceProps['source']}
                                        />
                                    </span>
                                </>
                            ) : (
                                'Never'
                            )}
                        </span>
                    </div>
                    <Text className="text-muted mt-1 mb-4">
                        <Icon aria-label="more-info text-normal" svgPath={mdiInformationOutline} /> The timestamp
                        indicates the last update made to the repository permissions of this user. This might have been
                        done via{' '}
                        <Link rel="noopener noreferrer" target="_blank" to="/help/admin/permissions/syncing">
                            permission syncing
                        </Link>{' '}
                        or{' '}
                        <Link rel="noopener noreferrer" target="_blank" to="/help/admin/permissions/api">
                            explicit permissions API
                        </Link>
                        . If the value <i>never</i> is displayed, it means we currently do not have any permission
                        records for the user. However, please note that the user may have had permissions stored in
                        Sourcegraph in the past.
                    </Text>
                    <ScheduleUserPermissionsSyncActionContainer user={user} />
                </>
            </Container>
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Sync Jobs' }]}
                description={
                    <>
                        List of permission sync jobs that fetch which <i>private</i> repositories the user can access on
                        the code host.
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
                description={
                    <>
                        List of <i>all</i> repositories the user can access on Sourcegraph.
                    </>
                }
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
    'Explicit API': {
        variant: 'success',
        tooltip: 'The permission was granted through explicit permissions API.',
    },
}

interface ScheduleUserPermissionsSyncActionContainerProps {
    user: { id: string; username: string }
}

class ScheduleUserPermissionsSyncActionContainer extends React.PureComponent<ScheduleUserPermissionsSyncActionContainerProps> {
    public render(): JSX.Element | null {
        return (
            <ActionContainer
                title="Manually schedule a permissions sync"
                description={
                    <div>
                        Schedules a high priority permissions sync job for the current user. This action will cancel all
                        previously queued user sync jobs.
                    </div>
                }
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

const permsSourceMap = {
    user_sync: 'user-centric permission sync',
    repo_sync: 'repo-centric permission sync',
    api: 'explicit permissions API',
}

interface PermissionSourceProps {
    source: 'user_sync' | 'repo_sync' | 'api' | null
}

const PermissionSource: React.FunctionComponent<PermissionSourceProps> = ({ source }) => {
    if (!source) {
        return <>unknown</>
    }
    return <>{permsSourceMap[source]}</>
}
