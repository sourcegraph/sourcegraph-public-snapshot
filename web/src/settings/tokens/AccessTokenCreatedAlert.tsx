import CircleCheckmarkIcon from '@sourcegraph/icons/lib/CircleCheckmark'
import * as React from 'react'
import { CopyableText } from '../../components/CopyableText'

interface AccessTokenCreatedAlertProps {
    className: string
    tokenSecret: string
}

/**
 * Displays a message informing the user to copy and save the newly created access token.
 */
export class AccessTokenCreatedAlert extends React.PureComponent<AccessTokenCreatedAlertProps> {
    public render(): JSX.Element | null {
        return (
            <div className={`access-token-created-alert ${this.props.className}`}>
                <p>
                    <CircleCheckmarkIcon className="icon-inline" /> Copy your new personal access token now. You won't
                    be able to see it again.
                </p>
                <CopyableText text={this.props.tokenSecret} size={48} />
                <h5 className="mt-4">
                    <strong>Example usage</strong>
                </h5>
                <pre className="mt-1">
                    <code>{curlExampleCommand(this.props.tokenSecret)}</code>
                </pre>
            </div>
        )
    }
}

function curlExampleCommand(tokenSecret: string): string {
    return `curl \\
  -H 'Authorization: token ${tokenSecret}' \\
  -d '{"query":"query { currentUser { username } }"}' \\
  ${window.context.appURL}/.api/graphql`
}
