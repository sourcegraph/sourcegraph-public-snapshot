import React from 'react'

import { Alert, Link } from '@sourcegraph/wildcard'

import { updateGitHubApp } from './UserAddCodeHostsPage'

export interface OrgUserNeedsGithubUpgrade {}

export const OrgUserNeedsGithubUpgrade: React.FunctionComponent<OrgUserNeedsGithubUpgrade> = () => (
    <Alert className="mb-4" role="alert" variant="warning">
        <h4>Update your code host connection with GitHub</h4>
        We’ve changed how we sync repositories with GitHub. Please{' '}
        <Link onClick={updateGitHubApp}>update your code host connection</Link> with GitHub to continue searching across
        your organization’s private repositories on Sourcegraph.
    </Alert>
)
