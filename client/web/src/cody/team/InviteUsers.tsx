import React, { useState, useCallback } from 'react'

import { pluralize } from '@sourcegraph/common'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { ButtonLink, H2, Link, Text, H3, TextArea } from '@sourcegraph/wildcard'

import { CodyAlert } from '../components/CodyAlert'
import { CodyContainer } from '../components/CodyContainer'
import { useSendInvite } from '../management/api/react-query/invites'
import { isValidEmailAddress } from '../util'

interface InviteUsersProps extends TelemetryV2Props {
    teamId: string
    remainingInviteCount: number
}

export const InviteUsers: React.FunctionComponent<InviteUsersProps> = ({
    teamId,
    remainingInviteCount,
    telemetryRecorder,
}) => {
    const [emailAddressesString, setEmailAddressesString] = useState<string>('')
    const emailAddresses = emailAddressesString.split(',').map(email => email.trim())
    const [emailAddressErrorMessage, setEmailAddressErrorMessage] = useState<string | null>(null)

    const sendInviteMutation = useSendInvite()

    const verifyEmailList = useCallback((): Error | void => {
        if (emailAddresses.length === 0) {
            return new Error('Please enter at least one email address.')
        }

        if (emailAddresses.length > remainingInviteCount) {
            return new Error(
                `${emailAddresses.length} email addresses entered, but you only have ${remainingInviteCount} seats.`
            )
        }

        const invalidEmails = emailAddresses.filter(email => !isValidEmailAddress(email))

        if (invalidEmails.length > 0) {
            return new Error(
                `Invalid email address${invalidEmails.length > 1 ? 'es' : ''}: ${invalidEmails.join(', ')}`
            )
        }
    }, [emailAddresses, remainingInviteCount])

    const onSendInvitesClicked = useCallback(async () => {
        const emailListError = verifyEmailList()
        if (emailListError) {
            setEmailAddressErrorMessage(emailListError.message)
            return
        }
        telemetryRecorder.recordEvent('cody.team.sendInvites', 'click', {
            metadata: { count: emailAddresses.length },
            privateMetadata: { teamId, emailAddresses },
        })

        try {
            await Promise.all(
                emailAddresses.map(emailAddress =>
                    sendInviteMutation.mutateAsync.call(undefined, { email: emailAddress, role: 'member' })
                )
            )

            telemetryRecorder.recordEvent('cody.team.sendInvites', 'success', {
                metadata: { count: emailAddresses.length },
                privateMetadata: { teamId, emailAddresses },
            })
        } catch (error) {
            telemetryRecorder.recordEvent('cody.team.sendInvites', 'error', {
                metadata: { count: emailAddresses.length, softError: 0 },
                privateMetadata: { teamId, emailAddresses, error },
            })
        }
    }, [emailAddresses, sendInviteMutation.mutateAsync, teamId, telemetryRecorder, verifyEmailList])

    return (
        <>
            {sendInviteMutation.status === 'success' && (
                <CodyAlert variant="greenSuccess">
                    <H3>
                        {emailAddresses.length} {pluralize('invite', emailAddresses.length)} sent!
                    </H3>
                    <Text size="small" className="mb-0">
                        Invitees will receive an email from cody@sourcegraph.com.
                    </Text>
                </CodyAlert>
            )}
            {sendInviteMutation.status === 'error' && (
                <CodyAlert variant="error">
                    <H3>Invites not sent.</H3>
                    <Text size="small" className="text-muted mb-0">
                        Error sending invites: {sendInviteMutation.error?.message}
                    </Text>
                    <Text size="small" className="mb-0">
                        If you encounter this issue repeatedly, please contact support at{' '}
                        <Link to="mailto:support@sourcegraph.com">support@sourcegraph.com</Link>.
                    </Text>
                </CodyAlert>
            )}

            <CodyContainer className="p-4 border bg-1 mb-4 d-flex flex-row">
                <div className="d-flex justify-content-between align-items-center w-100">
                    <div>
                        <img
                            src="https://storage.googleapis.com/sourcegraph-assets/cody/user-badges.png"
                            alt="User badges"
                            width="230"
                            height="202"
                            className="mr-3"
                        />
                    </div>
                    <div className="flex-1 d-flex flex-column">
                        <H2 className="mb-4 font-weight-normal">
                            <strong>Invite users</strong> – {remainingInviteCount}{' '}
                            {pluralize('seat', remainingInviteCount)} remaining
                        </H2>
                        <TextArea
                            className="mb-2"
                            placeholder="Example: someone@sourcegraph.com, another.user@sourcegraph.com"
                            rows={4}
                            onChange={event => {
                                setEmailAddressErrorMessage(null)
                                setEmailAddressesString(event.target.value)
                                sendInviteMutation.reset()
                            }}
                            isValid={emailAddressErrorMessage ? false : undefined}
                        />
                        {emailAddressErrorMessage ? (
                            <Text className="text-danger mb-2">{emailAddressErrorMessage}</Text>
                        ) : (
                            <Text className="text-muted mb-2">Enter email addresses separated by a comma.</Text>
                        )}

                        <div>
                            <ButtonLink variant="success" size="sm" onSelect={onSendInvitesClicked}>
                                Send
                            </ButtonLink>
                        </div>
                    </div>
                </div>
            </CodyContainer>
        </>
    )
}
