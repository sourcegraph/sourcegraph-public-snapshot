import React, { useState, useCallback } from 'react'

import classNames from 'classnames'

import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { ButtonLink, H2, Link, Text, H3, TextArea } from '@sourcegraph/wildcard'

import { isValidEmailAddress, fetchThroughSSCProxy } from '../util'

import styles from './CodyManageTeamPage.module.scss'

interface InviteUsersProps extends TelemetryV2Props {
    teamId: string | null
    remainingInviteCount: number
}

export const InviteUsers: React.FunctionComponent<InviteUsersProps> = ({
    teamId,
    remainingInviteCount,
    telemetryRecorder,
}) => {
    const [emailAddressesString, setEmailAddressesString] = useState<string>('')
    const [emailAddressErrorMessage, setEmailAddressErrorMessage] = useState<string | null>(null)
    const [invitesSendingStatus, setInvitesSendingStatus] = useState<'idle' | 'sending' | 'success' | 'error'>('idle')
    const [invitesSendingErrorMessage, setInvitesSendingErrorMessage] = useState<string | null>(null)

    const onSendInvitesClicked = useCallback(async () => {
        const {emails: emailAddresses, error: emailParsingError} = parseEmailList(emailAddressesString, remainingInviteCount)
        if (emailParsingError) {
            setEmailAddressErrorMessage(emailParsingError)
            return
        }
        telemetryRecorder.recordEvent('cody.team.sendInvites', 'click', {
            metadata: { count: emailAddresses.length },
            privateMetadata: { teamId, emailAddresses },
        })

        setInvitesSendingStatus('sending')
        try {
            const responses = await Promise.all(
                emailAddresses.map(emailAddress =>
                    fetchThroughSSCProxy('/team/current/invites', 'POST', { email: emailAddress, role: 'member' })
                )
            )
            if (responses.some(response => response.status !== 200)) {
                const responsesText = await Promise.all(responses.map(response => response.text()))
                setInvitesSendingStatus('error')
                setInvitesSendingErrorMessage(`Error sending invites: ${responsesText.join(', ')}`)
                telemetryRecorder.recordEvent('cody.team.sendInvites', 'error', {
                    metadata: { count: emailAddresses.length, softError: 1 },
                    privateMetadata: { teamId, emailAddresses },
                })

                return
            }
            setInvitesSendingStatus('success')
            telemetryRecorder.recordEvent('cody.team.sendInvites', 'success', {
                metadata: { count: emailAddresses.length },
                privateMetadata: { teamId, emailAddresses },
            })
        } catch (error) {
            setInvitesSendingStatus('error')
            setInvitesSendingErrorMessage(`Error sending invites: ${error}`)
            telemetryRecorder.recordEvent('cody.team.sendInvites', 'error', {
                metadata: { count: emailAddresses.length, softError: 0 },
                privateMetadata: { teamId, emailAddresses },
            })
        }
    }, [emailAddressesString, remainingInviteCount, teamId, telemetryRecorder])

    return (
        <>
            {invitesSendingStatus === 'success' && (
                <div className={classNames('mb-4', styles.alert, styles.blueSuccessAlert)}>
                    <H3>4 invites sent!</H3>
                    <Text size="small" className="mb-0">
                        Invitees will receive an email from cody@sourcegraph.com.
                    </Text>
                </div>
            )}
            {invitesSendingStatus === 'error' && (
                <div className={classNames('mb-4', styles.alert, styles.errorAlert)}>
                    <H3>Invites not sent.</H3>
                    <Text size="small" className="text-muted mb-0">
                        {invitesSendingErrorMessage}
                    </Text>
                    <Text size="small">
                        If you encounter this issue repeatedly, please contact support at{' '}
                        <Link to="mailto:support@sourcegraph.com">support@sourcegraph.com</Link>.
                    </Text>
                </div>
            )}

            {!!remainingInviteCount && (
                <div className={classNames('p-4 border bg-1 mb-4 d-flex flex-row', styles.container)}>
                    <div className="d-flex justify-content-between align-items-center w-100">
                        <div>
                            <img
                                src="https://storage.googleapis.com/sourcegraph-assets/cody/user-badges.png"
                                alt="User badges"
                                width="230"
                                height="202"
                                className={classNames('mr-3')}
                            />
                        </div>
                        <div className="flex-1 d-flex flex-column">
                            <H2 className={classNames('mb-4', styles.inviteUsersHeader)}>
                                <strong>Invite users</strong> – You have {remainingInviteCount} free seat
                                {remainingInviteCount === 1 ? '' : 's'}
                            </H2>
                            <TextArea
                                className={classNames('mb-2')}
                                placeholder="Example: someone@sourcegraph.com, another.user@sourcegraph.com"
                                rows={4}
                                onChange={event => {
                                    setEmailAddressErrorMessage(null)
                                    setEmailAddressesString(event.target.value)
                                }}
                            />
                            <Text className="text-muted mb-2">Enter email addresses separated by a comma.</Text>
                            <Text className="text-danger mb-2">{emailAddressErrorMessage}</Text>
                            <div className="d-flex justify-content-end">
                                <ButtonLink variant="success" size="sm" onSelect={onSendInvitesClicked}>
                                    Send
                                </ButtonLink>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </>
    )
}

function parseEmailList(emailAddressesString: string, remainingInviteCount: number) {
    const emails = emailAddressesString.split(',').map(email => email.trim())
    if (emails.length === 0) {
        return {emails, error: 'Please enter at least one email address.'}
    }

    if (emails.length > remainingInviteCount) {
        return {emails, error: `${emails.length} email addresses entered, but you only have ${remainingInviteCount} seats.`}
    }

    const invalidEmails = emails.filter(email => !isValidEmailAddress(email))

    if (invalidEmails.length > 0) {
        return {emails, error: `Invalid email address${invalidEmails.length > 1 ? 'es' : ''}: ${invalidEmails.join(', ')}`}
    }

    return {emails, error: null}
}
