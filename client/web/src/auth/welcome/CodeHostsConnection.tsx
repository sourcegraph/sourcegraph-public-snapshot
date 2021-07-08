import React from 'react'

import { UserCodeHosts } from '../../user/settings/codeHosts/UserCodeHosts'

interface CodeHostsConnection extends Omit<UserCodeHosts, 'onDidRemove' | 'onDidError'> {
    refetch: UserCodeHosts['onDidRemove']
}

export const CodeHostsConnection: React.FunctionComponent<CodeHostsConnection> = ({
    user,
    context,
    refetch,
    externalServices,
}) => (
    <>
        <div className="mb-4">
            <h3>Connect with code hosts</h3>
            <p className="text-muted">
                Connect with providers where your source code is hosted. Then, choose the repositories you’d like to
                search with Sourcegraph.
            </p>
        </div>
        <UserCodeHosts
            user={user}
            externalServices={externalServices}
            context={context}
            onDidError={error => console.warn('<UserCodeHosts .../>', error)}
            onDidRemove={() => refetch()}
        />
    </>
)
