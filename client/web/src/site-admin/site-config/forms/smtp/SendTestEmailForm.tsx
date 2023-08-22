import { FC, useCallback, useMemo, useState } from 'react'

import { FetchResult } from '@apollo/client'

import { useMutation } from '@sourcegraph/http-client'
import { Link, Alert, Button, Select, Label } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../auth'
import { LoaderButton } from '../../../../components/LoaderButton'
import {
    SendTestEmailToResult,
    SendTestEmailToVariables,
    SMTPAuthType,
    SMTPConfig,
} from '../../../../graphql-operations'
import { SEND_TEST_EMAIL } from '../../../backend'

import { FormData } from './index'

interface SendTestEmailProps {
    className?: string
    authenticatedUser: AuthenticatedUser
    formData: FormData
}
export const SendTestEmailForm: FC<SendTestEmailProps> = ({ authenticatedUser, className, formData }) => {
    const [email, setEmail] = useState('')

    const controller = useMemo(() => new AbortController(), [])

    const [sendTestEmail, { data, loading, error, reset }] = useMutation<
        SendTestEmailToResult,
        SendTestEmailToVariables
    >(SEND_TEST_EMAIL, {
        context: {
            fetchOptions: {
                signal: controller.signal,
            },
        },
    })
    const emails = useMemo(() => {
        if (authenticatedUser?.emails.length === 1) {
            setEmail(authenticatedUser.emails[0].email)
        }
        return authenticatedUser?.emails ?? []
    }, [authenticatedUser, setEmail])

    const emailChanged = useCallback(
        (event: React.ChangeEvent<HTMLSelectElement>) => {
            setEmail(event.target.value)
        },
        [setEmail]
    )

    const onSendTestEmail = useCallback((): Promise<FetchResult<SendTestEmailToResult>> => {
        let authentication = SMTPAuthType.NONE
        if (formData.authentication === 'PLAIN') {
            authentication = SMTPAuthType.PLAIN
        }
        if (formData.authentication === 'CRAM-MD5') {
            authentication = SMTPAuthType.CRAM_MD5
        }
        return sendTestEmail({
            variables: {
                to: email,
                config: {
                    ...formData,
                    authentication,
                } as SMTPConfig,
            },
        })
    }, [sendTestEmail, email, formData])

    const cancel = useCallback(() => {
        controller.abort()
        reset()
    }, [controller, reset])

    return (
        <div className={className}>
            {error && <Alert variant="danger">{error.message}</Alert>}
            {data && (
                <Alert variant={data.sendTestEmail.startsWith('Failed') ? 'danger' : 'success'}>
                    {data.sendTestEmail}
                </Alert>
            )}
            <Label className="w-100 mt-2" id="send-test-email-label">
                Send test email
            </Label>
            <Select
                aria-labelledby="send-test-email-label"
                name="authentication"
                message={
                    <>
                        Verify your configuration by choosing an email{' '}
                        <Link to={`${authenticatedUser.settingsURL}/emails`}>configured on your account</Link> to send a
                        test email.
                    </>
                }
                value={email}
                onChange={emailChanged}
            >
                <option key="empty" value="">
                    Choose email address
                </option>
                {emails.map(email => (
                    <option key={email.email} value={email.email}>
                        {email.email}
                        {email.isPrimary && ' (primary)'}
                    </option>
                ))}
            </Select>
            <div className="d-flex">
                <LoaderButton
                    onClick={onSendTestEmail}
                    loading={loading}
                    disabled={!email || loading}
                    label="Send test email"
                    variant="secondary"
                />
                <Button className="ml-2" onClick={cancel} variant="secondary">
                    Cancel
                </Button>
            </div>
        </div>
    )
}
