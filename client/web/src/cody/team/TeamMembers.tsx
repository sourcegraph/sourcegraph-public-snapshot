import { type FunctionComponent, useMemo, useCallback, useState } from 'react'

import classNames from 'classnames'

import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { H2, Text, Badge, Link, ButtonLink } from '@sourcegraph/wildcard'

import { fetchThroughSSCProxy } from '../util'

import styles from './CodyManageTeamPage.module.scss'

export interface TeamMember {
    accountId: string
    displayName: string | null
    email: string
    avatarUrl: string | null
    role: 'admin' | 'member'
}

export interface TeamInvite {
    id: string
    email: string
    role: 'admin' | 'member' | 'none'
    status: 'sent' | 'errored' | 'accepted' | 'canceled'
    error: string | null
    sentAt: string | null
    acceptedAt: string | null
}

interface TeamMemberListProps extends TelemetryV2Props {
    teamId: string | null
    teamMembers: TeamMember[]
    invites: TeamInvite[]
    isAdmin: boolean
}

export const TeamMemberList: FunctionComponent<TeamMemberListProps> = ({
    teamId,
    teamMembers,
    invites,
    isAdmin,
    telemetryRecorder,
}) => {
    const [loading, setLoading] = useState(false)
    const [actionResult, setActionResult] = useState<{ message: string; isError: boolean } | null>(null)
    const setActionResultWithTimeout = useCallback((message: string, isError: boolean) => {
        setLoading(false)
        setActionResult({ message, isError })
        setTimeout(() => {
            setActionResult(null)
        }, 5000)
    }, [])
    const setRole = useCallback(
        async (accountId: string, newRole: 'member' | 'admin'): Promise<void> => {
            if (!loading) {
                setLoading(true)
                telemetryRecorder.recordEvent('cody.team.revokeAdmin', 'click', {
                    privateMetadata: { teamId, accountId },
                })

                const response = await fetchThroughSSCProxy(
                    `/team/current/members/${accountId}?newRole=${newRole}`,
                    'PATCH'
                )
                if (!response.ok) {
                    setActionResultWithTimeout('Failed to revoke admin. Error code was: ' + response.status, true)
                } else {
                    setActionResultWithTimeout('Admin revoked.', false)
                }
            }
        },
        [loading, telemetryRecorder, teamId, setActionResultWithTimeout]
    )

    const revokeInvite = useCallback(
        async (inviteId: string): Promise<void> => {
            if (!loading) {
                setLoading(true)
                telemetryRecorder.recordEvent('cody.team.revokeInvite', 'click', { privateMetadata: { teamId } })

                const response = await fetchThroughSSCProxy(`/team/current/invites/${inviteId}/cancel`, 'POST')
                if (!response.ok) {
                    setActionResultWithTimeout('Failed to revoke invite. Error code was: ' + response.status, true)
                } else {
                    setActionResultWithTimeout('Invite revoked.', false)
                }
            }
        },
        [loading, telemetryRecorder, teamId, setActionResultWithTimeout]
    )

    const resendInvite = useCallback(
        async (inviteId: string): Promise<void> => {
            if (!loading) {
                setLoading(true)
                telemetryRecorder.recordEvent('cody.team.revokeInvite', 'click', { privateMetadata: { teamId } })

                const response = await fetchThroughSSCProxy(`/team/current/invites/${inviteId}/resend`, 'POST')
                if (!response.ok) {
                    setActionResultWithTimeout('Failed to resend invite. Error code was: ' + response.status, true)
                } else {
                    setActionResultWithTimeout('Invite resent.', false)
                }
            }

            telemetryRecorder.recordEvent('cody.team.resendInvite', 'click', { privateMetadata: { teamId } })
        },
        [loading, telemetryRecorder, teamId, setActionResultWithTimeout]
    )

    const removeMember = useCallback(
        async (accountId: string): Promise<void> => {
            telemetryRecorder.recordEvent('cody.team.removeMember', 'click', { privateMetadata: { teamId, accountId } })

            if (!loading) {
                setLoading(true)
                telemetryRecorder.recordEvent('cody.team.revokeInvite', 'click', { privateMetadata: { teamId } })

                const response = await fetchThroughSSCProxy(`/team/current/members/${accountId}`, 'DELETE')
                if (!response.ok) {
                    setActionResultWithTimeout('Failed to remove team member. Error code was: ' + response.status, true)
                } else {
                    setActionResultWithTimeout('Team member removed.', false)
                }
            }
        },
        [telemetryRecorder, teamId, loading, setActionResultWithTimeout]
    )

    const adminCount = useMemo(() => teamMembers?.filter(member => member.role === 'admin').length ?? 0, [teamMembers])

    if (!teamMembers) {
        return null
    }

    return (
        <>
            {actionResult && (
                <div
                    className={classNames(
                        'mb-4',
                        styles.alert,
                        actionResult.isError ? styles.errorAlert : styles.blueSuccessAlert
                    )}
                >
                    {actionResult.message}
                </div>
            )}
            <div className={classNames('p-4 border bg-1 d-flex flex-column', styles.container)}>
                <H2 className="text-lg font-semibold mb-2">Team members</H2>
                <Text className="text-sm text-gray-500 mb-4">Manage invited and active users</Text>
                <ul className="space-y-4 d-flex flex-column list-none pl-0">
                    {teamMembers.map(member => (
                        <li key={member.accountId} className="d-flex flex-row justify-between mb-4">
                            <div className="flex-1 d-flex flex-row">
                                {member.avatarUrl ? (
                                    <img
                                        src={member.avatarUrl}
                                        alt="avatar"
                                        width="40"
                                        height="40"
                                        className={classNames(styles.avatar)}
                                    />
                                ) : (
                                    <div className={classNames(styles.avatar, styles.avatarPlaceholder)} />
                                )}
                                <div className="d-flex flex-column justify-content-center ml-2">
                                    {member.displayName && <strong>{member.displayName}</strong>}
                                    <Text className="mb-0">{member.email}</Text>
                                </div>
                                {member.role === 'admin' && (
                                    <div className="d-flex flex-column justify-content-center ml-2">
                                        <Badge variant="primary">ADMIN</Badge>
                                    </div>
                                )}
                            </div>
                            {isAdmin && (
                                <div className="d-flex">
                                    {member.role === 'admin' ? (
                                        <div className="d-flex flex-column justify-content-center ml-2">
                                            <Link
                                                to="#"
                                                onClick={() => setRole(member.accountId, 'member')}
                                                className="ml-2"
                                                aria-disabled={adminCount < 2}
                                            >
                                                Revoke admin
                                            </Link>
                                        </div>
                                    ) : (
                                        <>
                                            <div className="d-flex flex-column justify-content-center ml-2">
                                                <Link
                                                    to="#"
                                                    onClick={() => setRole(member.accountId, 'admin')}
                                                    className="ml-2"
                                                >
                                                    Make admin
                                                </Link>
                                            </div>
                                            <div className="d-flex flex-column justify-content-center ml-2">
                                                <ButtonLink
                                                    to="#"
                                                    variant="danger"
                                                    size="sm"
                                                    onClick={() => removeMember(member.accountId)}
                                                    className="ml-2"
                                                >
                                                    Remove
                                                </ButtonLink>
                                            </div>
                                        </>
                                    )}
                                </div>
                            )}
                        </li>
                    ))}
                    {invites
                        .filter(invite => invite.status === 'sent')
                        .map(invite => (
                            <li key={invite.id} className="d-flex flex-row justify-between mb-4">
                                <div className="flex-1 d-flex flex-row">
                                    <div className={classNames(styles.avatar, styles.avatarPlaceholder)} />
                                    <div className="d-flex flex-column justify-content-center ml-2">
                                        <Text className="mb-0">{invite.email}</Text>
                                    </div>
                                    {invite.role === 'admin' && (
                                        <div className="d-flex flex-column justify-content-center ml-2">
                                            <Badge variant="primary">ADMIN</Badge>
                                        </div>
                                    )}
                                    <div className="d-flex flex-column justify-content-center ml-2">
                                        <div className="d-flex flex-row">
                                            <div className="d-flex flex-column justify-content-center ml-2">
                                                <Badge variant="secondary">INVITED</Badge>
                                            </div>
                                            <em className="ml-4">Invite sent {invite.sentAt /* TODO format this */}</em>
                                        </div>
                                    </div>
                                </div>
                                {isAdmin && (
                                    <div className="d-flex">
                                        <div className="d-flex row justify-content-center ml-2">
                                            <div className="d-flex flex-column justify-content-center ml-2">
                                                <Link to="#" onClick={() => revokeInvite(invite.id)} className="ml-2">
                                                    Revoke
                                                </Link>
                                            </div>
                                            <div className="d-flex flex-column justify-content-center ml-2">
                                                <ButtonLink
                                                    to="#"
                                                    variant="success"
                                                    size="sm"
                                                    onClick={() => resendInvite(invite.id)}
                                                    className="ml-2"
                                                >
                                                    Resend invite
                                                </ButtonLink>
                                            </div>
                                        </div>
                                    </div>
                                )}
                            </li>
                        ))}
                </ul>
            </div>
        </>
    )
}
