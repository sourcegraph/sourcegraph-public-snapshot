import classNames from 'classnames'
import React from 'react'

import { CodeSnippet } from '@sourcegraph/branded/src/components/CodeSnippet'

import { AccessTokenScopes } from '../../auth/accessToken'
import { CopyableText } from '../../components/CopyableText'
import { AccessTokenFields } from '../../graphql-operations'

interface AccessTokenCreatedAlertProps {
    token: AccessTokenFields
    tokenSecret: string
    className?: string
}

/**
 * Displays a message informing the user to copy and save the newly created access token.
 */
export const AccessTokenCreatedAlert: React.FunctionComponent<AccessTokenCreatedAlertProps> = ({
    token,
    tokenSecret,
    className,
}) => {
    const isSudoToken = token.scopes.includes(AccessTokenScopes.SiteAdminSudo)
    return (
        <div className={classNames('access-token-created-alert alert alert-success', className)}>
            <p>Copy the new access token now. You won't be able to see it again.</p>
            <CopyableText className="test-access-token" text={tokenSecret} size={48} />
            <h5 className="mt-4 mb-2">
                <strong>Example usage</strong>
            </h5>
            <CodeSnippet code={curlExampleCommand(tokenSecret, isSudoToken)} className="mb-0" language="bash" />
        </div>
    )
}

function curlExampleCommand(tokenSecret: string, isSudoToken: boolean): string {
    const credentials = isSudoToken
        ? `token-sudo user="SUDO-TO-USERNAME",token="${tokenSecret}"`
        : `token ${tokenSecret}`

    return `curl \\
  -H 'Authorization: ${credentials}' \\
  -d '{"query":"query { currentUser { username } }"}' \\
  ${window.context.externalURL}/.api/graphql`
}
