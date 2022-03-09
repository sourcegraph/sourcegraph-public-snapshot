import React from 'react'

import { Container, Button, Link } from '@sourcegraph/wildcard'

export interface OrgUserNeedsGithubUpgrade {
    orgDisplayName: string
    user: {
        id: string
        username: string
    }
}

export const OrgUserNeedsGithubUpgrade: React.FunctionComponent<OrgUserNeedsGithubUpgrade> = ({
    user,
    orgDisplayName,
}) => (
    <Container className="mb-4">
        <p>
            Upgrade your GitHub code-host connection to start searching across the {orgDisplayName} organization's
            private repositories on Sourcegraph.
        </p>
        <Button to={`/users/${user.username}/settings/code-hosts`} variant="primary" as={Link}>
            Connect with Github
        </Button>
    </Container>
)
