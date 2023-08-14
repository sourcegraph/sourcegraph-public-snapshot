import { FC, useCallback, useMemo } from 'react'

import { useMutation } from '@sourcegraph/http-client'
import { Link, Alert, Text, Button } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { LoaderButton } from '../../components/LoaderButton'
import { SendTestEmailToResult, SendTestEmailToVariables } from '../../graphql-operations'
import { SEND_TEST_EMAIL } from '../backend'

interface SendTestEmailProps {
    className?: string
    authenticatedUser: AuthenticatedUser
}
export const SendTestEmailForm: FC<SendTestEmailProps> = ({ authenticatedUser, className }) => {
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
        onError: err => console.error('Apollo error', err),
    })
    const primaryEmail = useMemo(
        () => authenticatedUser?.emails.find(email => email.isPrimary)?.email,
        [authenticatedUser]
    )
    const onSendTestEmail = useCallback(() => {
        sendTestEmail({
            variables: {
                to: primaryEmail!,
            },
        })
    }, [sendTestEmail, primaryEmail])

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
            <Text>
                Verify currently saved configuration by sending an email to your primary email address (
                <code>{primaryEmail}</code>) configured on{' '}
                <Link to={`${authenticatedUser.settingsURL}/emails`}>your Sourcegraph account</Link>.
            </Text>
            <div className="w-100 d-flex justify-content-end">
                <LoaderButton
                    onClick={onSendTestEmail}
                    loading={loading}
                    disabled={!primaryEmail || loading}
                    label="Send test email"
                    variant="primary"
                />
                <Button className="ml-2" onClick={cancel} variant="secondary">
                    Cancel
                </Button>
            </div>
        </div>
    )
}
