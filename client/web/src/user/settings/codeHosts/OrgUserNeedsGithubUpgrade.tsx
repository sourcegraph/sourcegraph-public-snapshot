import React from 'react'

import { Alert, ButtonLink } from '@sourcegraph/wildcard'

import { updateGitHubApp } from './UserAddCodeHostsPage'

export interface OrgUserNeedsGithubUpgrade {}

export const OrgUserNeedsGithubUpgrade: React.FunctionComponent<OrgUserNeedsGithubUpgrade> = () => (
    <Alert className="mb-4" role="alert" variant="warning">
        <h4>Update your code host connection with GitHub</h4>
        We’ve changed how we sync repositories with GitHub. Please{' '}
        <ButtonLink onClick={updateGitHubApp} variant="link" display="inline" className="align-baseline m-0 p-0">
            update your code host connection
        </ButtonLink>{' '}
        with GitHub to continue searching across your organization’s private repositories on Sourcegraph.
    </Alert>
)
