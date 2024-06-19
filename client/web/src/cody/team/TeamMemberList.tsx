import { type FunctionComponent, useMemo, useCallback, useState } from 'react'

import classNames from 'classnames'
import { intlFormatDistance } from 'date-fns'

import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { H2, Text, Badge, Link, ButtonLink } from '@sourcegraph/wildcard'

import { CodyAlert } from '../components/CodyAlert'
import { CodyContainer } from '../components/CodyContainer'
import { useCancelInvite, useResendInvite } from '../management/api/react-query/invites'
import { useUpdateTeamMember } from '../management/api/react-query/teams'
import type { TeamMember, TeamInvite } from '../management/api/types'

import styles from './TeamMemberList.module.scss'

interface TeamMemberListProps extends TelemetryV2Props {
    teamId: string
    teamMembers: TeamMember[]
    invites: Omit<TeamInvite, 'sentBy'>[]
    isAdmin: boolean
}

// This tiny function is extracted to make it testable. Same for the "now" parameter.
export const formatInviteDate = (sentAt: string | null, now?: Date): string => {
    try {
        if (sentAt) {
            return intlFormatDistance(sentAt, now ?? new Date())
        }
        return ''
    } catch {
        return ''
    }
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
    const updateTeamMemberMutation = useUpdateTeamMember()
    const cancelInviteMutation = useCancelInvite()
    const resendInviteMutation = useResendInvite()
    const updateRole = useCallback(
        async (accountId: string, newRole: 'member' | 'admin'): Promise<void> => {
            if (!loading) {
                // Avoids sending multiple requests at once
                setLoading(true)
                telemetryRecorder.recordEvent('cody.team.revokeAdmin', 'click', {
                    privateMetadata: { teamId, accountId },
                })

                try {
                    await updateTeamMemberMutation.mutateAsync.call(undefined, {
                        updateMemberRole: { accountId, teamRole: newRole },
                    })
                    setLoading(false)
                    setActionResult({ message: 'Team role updated.', isError: false })
                } catch (error) {
                    setLoading(false)
                    setActionResult({
                        message: `We couldn't modify the user's role. The error was: "${error}". Please try again later.`,
                        isError: true,
                    })
                }
            }
        },
        [loading, telemetryRecorder, teamId, updateTeamMemberMutation.mutateAsync]
    )

    const revokeInvite = useCallback(
        async (inviteId: string): Promise<void> => {
            if (!loading) {
                // Avoids sending multiple requests at once
                setLoading(true)
                telemetryRecorder.recordEvent('cody.team.revokeInvite', 'click', { privateMetadata: { teamId } })

                try {
                    await cancelInviteMutation.mutateAsync.call(undefined, { teamId, inviteId })
                    setLoading(false)
                    setActionResult({ message: 'Invite revoked.', isError: false })
                } catch (error) {
                    setLoading(false)
                    setActionResult({
                        message: `We couldn't revoke the invite. The error was: "${error}". Please try again later.`,
                        isError: true,
                    })
                }
            }
        },
        [loading, telemetryRecorder, teamId, cancelInviteMutation.mutateAsync]
    )

    const resendInvite = useCallback(
        async (inviteId: string): Promise<void> => {
            if (!loading) {
                // Avoids sending multiple requests at once
                setLoading(true)
                telemetryRecorder.recordEvent('cody.team.resendInvite', 'click', { privateMetadata: { teamId } })

                try {
                    await resendInviteMutation.mutateAsync.call(undefined, { inviteId })
                    setLoading(false)
                    setActionResult({ message: 'Invite resent.', isError: false })
                } catch (error) {
                    setLoading(false)
                    setActionResult({
                        message: `We couldn't resend the invite (${error}). Please try again later.`,
                        isError: true,
                    })
                }
            }

            telemetryRecorder.recordEvent('cody.team.resendInvite', 'click', { privateMetadata: { teamId } })
        },
        [loading, telemetryRecorder, teamId, resendInviteMutation.mutateAsync]
    )

    const removeMember = useCallback(
        async (accountId: string): Promise<void> => {
            if (!loading) {
                setLoading(true)
                telemetryRecorder.recordEvent('cody.team.removeMember', 'click', { privateMetadata: { teamId } })

                try {
                    await updateTeamMemberMutation.mutateAsync.call(undefined, {
                        removeMember: { accountId, teamRole: 'member' },
                    })
                    setLoading(false)
                    setActionResult({ message: 'Team member removed.', isError: false })
                } catch (error) {
                    setLoading(false)
                    setActionResult({
                        message: `We couldn't remove the team member. (${error}). Please try again later.`,
                        isError: true,
                    })
                }
            }
        },
        [loading, telemetryRecorder, teamId, updateTeamMemberMutation.mutateAsync]
    )

    const adminCount = useMemo(() => teamMembers?.filter(member => member.role === 'admin').length ?? 0, [teamMembers])

    if (!teamMembers) {
        return null
    }

    return (
        <>
            {actionResult && (
                <CodyAlert variant={actionResult.isError ? 'error' : 'greenSuccess'}>{actionResult.message}</CodyAlert>
            )}
            <CodyContainer className={classNames('p-4 border bg-1 d-flex flex-column')}>
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
            </CodyContainer>
        </>
    )
}
