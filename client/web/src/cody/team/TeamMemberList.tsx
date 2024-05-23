import { type FunctionComponent, useMemo, useCallback, useState } from 'react'

import classNames from 'classnames'
import { intlFormatDistance } from 'date-fns'

import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { H2, Text, Badge, Link, ButtonLink } from '@sourcegraph/wildcard'

import { requestSSC } from '../util'

import styles from './CodyManageTeamPage.module.scss'

export interface TeamMember {
    accountId: string
    displayName: string | null
    email: string
    avatarUrl: string | null
    role: 'admin' | 'member'
}

interface TeamMemberListProps extends TelemetryV2Props {
    teamId: string | null
    teamMembers: TeamMember[]
    invites: TeamInvite[]
    isAdmin: boolean
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

// This tiny function is extracted to make it testable. Same for the "now" parameter.
export const formatInviteDate = (sentAt: string | null, now?: Date): string => {
    try {
        if (sentAt) {
            return intlFormatDistance(sentAt || '', now ?? new Date())
        }
    } catch {
        return ''
    }
    return ''
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
    const updateRole = useCallback(
        async (accountId: string, newRole: 'member' | 'admin'): Promise<void> => {
            if (!loading) {
                // Avoids sending multiple requests at once
                setLoading(true)
                telemetryRecorder.recordEvent('cody.team.revokeAdmin', 'click', {
                    privateMetadata: { teamId, accountId },
                })

                try {
                    const response = await requestSSC(`/team/current/members/${accountId}?newRole=${newRole}`, 'PATCH')
                    if (!response.ok) {
                        setLoading(false)
                        setActionResult({
                            message: `We couldn't modify the user's role (${response.status}). Please try again later.`,
                            isError: true,
                        })
                    } else {
                        setLoading(false)
                        setActionResult({ message: 'Team role updated.', isError: false })
                    }
                } catch (error) {
                    setLoading(false)
                    setActionResult({
                        message: `We couldn't modify the user's role. The error was: "${error}". Please try again later.`,
                        isError: true,
                    })
                }
            }
        },
        [loading, telemetryRecorder, teamId]
    )

    const revokeInvite = useCallback(
        async (inviteId: string): Promise<void> => {
            if (!loading) {
                // Avoids sending multiple requests at once
                setLoading(true)
                telemetryRecorder.recordEvent('cody.team.revokeInvite', 'click', { privateMetadata: { teamId } })

                const response = await requestSSC(`/team/current/invites/${inviteId}/cancel`, 'POST')
                if (!response.ok) {
                    setLoading(false)
                    setActionResult({
                        message: `We couldn't revoke the invite (${response.status}). Please try again later.`,
                        isError: true,
                    })
                } else {
                    setLoading(false)
                    setActionResult({ message: 'Invite revoked.', isError: false })
                }
            }
        },
        [loading, telemetryRecorder, teamId]
    )

    const resendInvite = useCallback(
        async (inviteId: string): Promise<void> => {
            if (!loading) {
                // Avoids sending multiple requests at once
                setLoading(true)
                telemetryRecorder.recordEvent('cody.team.resendInvite', 'click', { privateMetadata: { teamId } })

                const response = await requestSSC(`/team/current/invites/${inviteId}/resend`, 'POST')
                if (!response.ok) {
                    setLoading(false)
                    setActionResult({
                        message: `We couldn't resend the invite (${response.status}). Please try again later.`,
                        isError: true,
                    })
                } else {
                    setLoading(false)
                    setActionResult({ message: 'Invite resent.', isError: false })
                }
            }

            telemetryRecorder.recordEvent('cody.team.resendInvite', 'click', { privateMetadata: { teamId } })
        },
        [loading, telemetryRecorder, teamId]
    )

    const removeMember = useCallback(
        async (accountId: string): Promise<void> => {
            if (!loading) {
                setLoading(true)
                telemetryRecorder.recordEvent('cody.team.removeMember', 'click', { privateMetadata: { teamId } })

                const response = await requestSSC(`/team/current/members/${accountId}`, 'DELETE')
                if (!response.ok) {
                    setLoading(false)
                    setActionResult({
                        message: `We couldn't remove the team member. (${response.status}). Please try again later.`,
                        isError: true,
                    })
                } else {
                    setLoading(false)
                    setActionResult({ message: 'Team member removed.', isError: false })
                }
            }
        },
        [telemetryRecorder, teamId, loading]
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
                <ul className={classNames(styles.teamMemberList, 'list-none pl-0')}>
                    {teamMembers.map(member => (
                        <li key={member.accountId} className="d-contents">
                            <div className="align-content-center">
                                <div className="d-flex flex-row">
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
                                </div>
                            </div>
                            <div className="align-content-center">
                                {member.role === 'admin' && (
                                    <Badge variant="primary" className="text-uppercase">
                                        admin
                                    </Badge>
                                )}
                            </div>
                            <div />
                            {isAdmin ? (
                                member.role === 'admin' ? (
                                    <>
                                        <div />
                                        <div className="align-content-center text-center">
                                            <Link
                                                to="#"
                                                onClick={() => updateRole(member.accountId, 'member')}
                                                className="ml-2"
                                                aria-disabled={adminCount < 2}
                                            >
                                                Revoke admin
                                            </Link>
                                        </div>
                                    </>
                                ) : (
                                    <>
                                        <div className="align-content-center text-center">
                                            <Link
                                                to="#"
                                                onClick={() => updateRole(member.accountId, 'admin')}
                                                className="ml-2"
                                            >
                                                Make admin
                                            </Link>
                                        </div>
                                        <div className="align-content-center text-center">
                                            <Link
                                                to="#"
                                                onClick={() => removeMember(member.accountId)}
                                                className="ml-2"
                                            >
                                                Remove
                                            </Link>
                                        </div>
                                    </>
                                )
                            ) : (
                                <>
                                    <div />
                                    <div />
                                </>
                            )}
                        </li>
                    ))}
                    {invites
                        .filter(invite => invite.status === 'sent')
                        .map(invite => (
                            <li key={invite.id} className="d-contents">
                                <div className="align-content-center">
                                    <div className="d-flex flex-row">
                                        <div className={classNames(styles.avatar, styles.avatarPlaceholder)} />
                                        <div className="d-flex flex-column justify-content-center ml-2">
                                            <Text className="mb-0">{invite.email}</Text>
                                        </div>
                                    </div>
                                </div>
                                <div className="align-content-center">
                                    <Badge variant="secondary" className="mr-2 text-uppercase">
                                        invited
                                    </Badge>
                                    {invite.role === 'admin' && (
                                        <Badge variant="primary" className="text-uppercase">
                                            admin
                                        </Badge>
                                    )}
                                </div>
                                <div className="align-content-center">
                                    <em>Invite sent {formatInviteDate(invite.sentAt)}</em>
                                </div>
                                {isAdmin && (
                                    <>
                                        <div className="align-content-center text-center">
                                            <Link to="#" onClick={() => revokeInvite(invite.id)} className="ml-2">
                                                Revoke
                                            </Link>
                                        </div>
                                        <div className="align-content-center text-center">
                                            <ButtonLink
                                                to="#"
                                                variant="secondary"
                                                size="sm"
                                                onClick={() => resendInvite(invite.id)}
                                                className="ml-2"
                                            >
                                                Re-send invite
                                            </ButtonLink>
                                        </div>
                                    </>
                                )}
                            </li>
                        ))}
                </ul>
            </div>
        </>
    )
}
