import React from 'react'

import { Text, Alert, Link } from '@sourcegraph/wildcard'

import { CopyableText } from '../../components/CopyableText'

interface AccountCreatedAlertProps {
    username: string
    email?: string
    resetPasswordURL?: string | null
    showEmail?: boolean | null
}

/**
 * An alert component that displays a message indicating that an account has been created.
 * - Shows a reset password URL in Copyable text if one is provided.
 * - Shows a link to the user's profile page.
 * - Shows a different message depending on whether email provided and email sending is enabled.
 */
export const AccountCreatedAlert: React.FunctionComponent<React.PropsWithChildren<AccountCreatedAlertProps>> = ({
    username,
    email,
    resetPasswordURL,
    children,
    showEmail,
}) => (
    <Alert variant="success">
        <Text>
            {showEmail ? (
                <>Account created for {email}.</>
            ) : (
                <>
                    Account created for <Link to={`/users/${username}`}>{username}</Link>.
                </>
            )}
        </Text>
        <Text>
            {resetPasswordURL
                ? window.context.emailEnabled && email
                    ? "A password reset URL has been sent to the new user's email address. If they don't receive it, you can also share the following password reset link: "
                    : "'You must manually send this password reset link to the new user: '"
                : 'The user must authenticate using a configured authentication provider.'}
        </Text>
        {resetPasswordURL && <CopyableText text={resetPasswordURL} size={40} />}
        {children}
    </Alert>
)
